# Deploying parser-app to Azure, step by step

You run every command yourself. Each step explains:

- **What** it creates.
- **Why** this architecture needs it, and why this Azure service rather than another.
- **Trade-offs**: what you give up, and the alternative.
- **Verify**: how to prove the step worked before moving on. Don't skip these. A broken step
  found later costs far more time to trace back.

Budget about 3–4 hours for the first pass. Redis alone takes ~20 minutes to provision, and you
can keep working while it does.

---

## 0. What changed in the code, and a local rehearsal

The app could not run on Azure as it was. Branch `azure-deploy` adds:

| Change | Where | Why Azure needs it |
|---|---|---|
| Storage interface with MinIO and Azure Blob backends, picked by `STORAGE_BACKEND` | `apps/api/internal/storage/`, `apps/worker/app/storage/` | Blob Storage does not speak the S3 protocol the app used |
| SAS upload URLs: create+write only, one blob, HTTPS only, 1 h expiry | `apps/api/internal/storage/azure.go` | Azure's equivalent of an S3 presigned URL. The account key never leaves the API |
| `upload_headers` in the upload response; the UI and e2e script send them | `upload_service.go`, `App.tsx`, `scripts/e2e-upload.sh` | Azure rejects an upload without `x-ms-blob-type: BlockBlob`. MinIO needs no headers, so it returns none |
| `REDIS_TLS=true` with certificate verification | API `main.go`, worker `main.py` | Azure Cache for Redis only accepts TLS (port 6380) |
| Worker finishes its current job on SIGTERM, then exits | worker `main.py` | Container Apps sends SIGTERM when it scales the worker to zero. Before, that killed jobs mid-flight |
| `API_KEY` is enforced on `/api/v1/*` when `AUTH_ENABLED=true` | `middleware/apikey.go`, `App.tsx` | Before, the variable was ignored and any caller could search the data or send `X-User-Roles: admin` |
| Azure SDK logging quieted to WARNING | worker `logging_config.py` | Every HTTP call was logged at INFO, which costs money in Log Analytics |
| `staticwebapp.config.json` | `apps/web/public/` | Lets Static Web Apps serve deep links (`/jobs/123`) from `index.html` |
| `docker-compose.azurite.yml` | repo root | Runs the Azure Blob code path locally against Microsoft's emulator |
| `.github/workflows/azure-deploy.yml` | | Repeatable deploys after the first manual one (section 14) |

MinIO stays the default, so `make up` and the Render blueprint behave exactly as before.

**Rehearse locally first.** This catches code problems for free, before you spend any credit:

```bash
cd parser
docker compose -f docker-compose.yml -f docker-compose.azurite.yml up -d --build
docker logs file-platform-api 2>&1 | grep "object storage ready"   # "backend":"azure"
make e2e-upload                                                    # must end "passed"
docker compose up -d --remove-orphans                              # back to MinIO
```

---

## 1. Architecture and the decisions behind it

```
  You (browser)                                     GitHub Actions (section 14)
      │                                                     │ OIDC, no stored password
      ├──▶ Static Web Apps ── React UI (CDN, free TLS)      ▼
      │                                         Container Registry (images)
      │ HTTPS + X-API-Key, only from your IP                │ AcrPull via managed identity
      ▼                                                     ▼
 ┌──────────── Container Apps environment ─ VNet 10.20.0.0/16, subnet snet-aca ─────────────┐
 │  ca-parser-api     Go · external ingress · IP allow-list · scale 0→2                    │
 │  ca-parser-worker  Python · no ingress · KEDA: 0→2 replicas by Redis list length       │
 └──────┬──────────────────┬──────────────────────┬────────────────────────┬─────────────────┘
        │ TLS               │ TLS :6380           │ private IP :9200      │ HTTPS
        ▼                   ▼                     ▼                        ▼
  PostgreSQL Flexible   Azure Cache for     VM (snet-search)        Blob Storage
  B1ms                  Redis C0            OpenSearch in Docker    ▲ browser PUTs the file
                                                                     │ directly with a SAS URL
  Secrets: Key Vault ◀── read by the user-assigned managed identity id-parser
  Logs: Log Analytics workspace (0.5 GB/day cap)
```

### Why each service

| Need | Chosen | Alternatives considered | Why this one |
|---|---|---|---|
| Run API and worker containers | **Container Apps** | App Service; AKS; a VM with docker compose | Serverless containers with scale-to-zero and built-in KEDA queue scaling. AKS would teach Kubernetes but costs ~$70+/month just for the control plane and nodes. App Service has no queue-driven scale-to-zero |
| Host the UI | **Static Web Apps (Free)** | Blob static website + Front Door; an nginx container | Free, global CDN, TLS, SPA routing. A container would bill 24/7 for static files |
| Images | **Container Registry (Basic)** | Docker Hub, GHCR | Pulls stay inside Azure, and authentication uses the managed identity, so the registry needs no password |
| Database | **PostgreSQL Flexible Server, B1ms** | Postgres in a container; Azure SQL | Managed backups, patching, TLS. B1ms is covered by the free-account 750 h/month for 12 months. The app is Postgres-specific (pgcrypto, JSONB) |
| Queue | **Azure Cache for Redis, C0** | Redis container; Service Bus | The code uses Redis lists, and KEDA has a Redis scaler. Service Bus is a more "Azure-native" queue but would mean rewriting the producer and consumer |
| Files | **Blob Storage** | MinIO on Azure Files | Native, cheapest (~$0.02/GB), SAS URLs give direct browser upload. MinIO on an SMB share is fragile |
| Search | **OpenSearch on a B2s VM** | Container App; Aiven/Elastic Cloud; Azure AI Search | Azure has no managed OpenSearch. A Container App with 4 GiB always on costs ~$150/month. Azure AI Search would mean a rewrite. The VM is the cheapest Azure-hosted option and teaches VMs, NSGs and VNets |
| Secrets | **Key Vault + managed identity** | Plain Container Apps secrets | One place to rotate. Apps authenticate as an identity, so no secret is needed to fetch secrets |
| Logs | **Log Analytics** | None | Container stdout and KEDA scaling events, queryable with KQL |

### Monthly cost (approximate, Pay-As-You-Go list prices)

| Item | $/month | Notes |
|---|---|---|
| OpenSearch VM B2s + 30 GB disk | ~35 | Deallocate it when idle (section 15) |
| Azure Cache for Redis C0 | ~16 | Cannot be paused |
| Container Registry Basic | 5 | |
| PostgreSQL B1ms + 32 GB | 0–15 | Free-account allowance for 12 months |
| Container Apps | 0–10 | Monthly free grant: 180k vCPU-s and 360k GiB-s. Scale-to-zero keeps you inside it |
| Static Web Apps, Key Vault, Blob, Log Analytics | ~1–3 | |
| **Total** | **~60–85** | Your $200 Free Trial credit covers the whole 30-day trial |

Your subscription is a **Free Trial with a spending limit**. When the credit runs out, Azure
disables resources rather than charging your card. You can't overspend by accident.

---

## 2. Prerequisites and shell variables

```bash
az login
az account show -o table                       # "Azure subscription 1", FreeTrial
az extension add --name containerapp --upgrade

for ns in Microsoft.App Microsoft.OperationalInsights Microsoft.ContainerRegistry \
          Microsoft.DBforPostgreSQL Microsoft.Cache Microsoft.KeyVault Microsoft.Web \
          Microsoft.Storage Microsoft.Network Microsoft.Compute Microsoft.ManagedIdentity; do
  az provider register --namespace $ns
done
```

**Why:** Each Azure service is a *resource provider* with its own API. New subscriptions must
register a provider before using it. Registration is free and takes a minute or two.

Now set variables. **Run every later step in this same terminal.** If you open a new one,
re-run this block with the same `SUFFIX`.

```bash
cd ~/mastering/chanchal-soc/pp-main/parser     # repo root: commands use relative paths
export LOCATION=centralindia     # one region for everything (latency + no cross-region egress)
export RG=rg-parser
export SUFFIX=12345              # pick 4–6 digits once; it makes global names unique
export ACR=acrparser$SUFFIX      # global names: lowercase letters and digits
export KV=kv-parser-$SUFFIX
export STORAGE=stparser$SUFFIX
export PG=pg-parser-$SUFFIX
export REDIS=redis-parser-$SUFFIX
export MYIP=$(curl -s https://api.ipify.org)
echo "$MYIP"
```

**Trade-off on region:** Choose one close to you. Some free-trial regions don't offer every SKU
(B-series VMs and Redis are the usual gaps). If a create fails with `SkuNotAvailable`, delete
the resource group and start again in another region, such as `westeurope` or `eastus`.

---

## 3. Resource group and logging

```bash
az group create -n $RG -l $LOCATION

az monitor log-analytics workspace create -g $RG -n log-parser -l $LOCATION
az monitor log-analytics workspace update -g $RG -n log-parser --quota 0.5
export LOG_ID=$(az monitor log-analytics workspace show -g $RG -n log-parser --query customerId -o tsv)
export LOG_KEY=$(az monitor log-analytics workspace get-shared-keys -g $RG -n log-parser --query primarySharedKey -o tsv)
```

**What and why:** A resource group is the lifecycle boundary for everything in this guide. You
can inspect, cost-filter and delete it all as one unit. The Log Analytics workspace receives every
container's stdout/stderr and the platform's system events.

**Trade-off:** The 0.5 GB/day cap means that during a log storm Azure drops logs instead of billing
you. That's fine for learning. In production you would alert on the cap instead.

**Verify:** `az resource list -g $RG -o table` lists `log-parser`.

---

## 4. Network

```bash
az network vnet create -g $RG -n vnet-parser -l $LOCATION \
  --address-prefixes 10.20.0.0/16 \
  --subnet-name snet-aca --subnet-prefixes 10.20.0.0/23
az network vnet subnet update -g $RG --vnet-name vnet-parser -n snet-aca \
  --delegations Microsoft.App/environments
az network vnet subnet create -g $RG --vnet-name vnet-parser -n snet-search \
  --address-prefixes 10.20.4.0/24
export ACA_SUBNET_ID=$(az network vnet subnet show -g $RG --vnet-name vnet-parser -n snet-aca --query id -o tsv)
```

**What:** A private network with two subnets: one handed to Container Apps, one for the OpenSearch VM.

**Why:** OpenSearch runs with no authentication, the same as locally. It must therefore be
reachable **only** from the apps, never from the internet. Putting the apps and the VM in one VNet
lets them talk over private IPs. *Delegation* gives `snet-aca` to the Container Apps service,
which places its infrastructure there.

**Trade-off:** Postgres, Redis and Blob are still reached over their public endpoints (TLS and
password or key). Making them private too means adding Private Endpoints and Private DNS zones:
roughly $7/month each and many more moving parts. That's a good next exercise once this works.

---

## 5. Identity and Key Vault

```bash
az identity create -g $RG -n id-parser
export ID_RES=$(az identity show -g $RG -n id-parser --query id -o tsv)
export ID_PRINCIPAL=$(az identity show -g $RG -n id-parser --query principalId -o tsv)

az keyvault create -g $RG -n $KV -l $LOCATION --enable-rbac-authorization true
export KV_ID=$(az keyvault show -n $KV --query id -o tsv)
export KV_URI=$(az keyvault show -n $KV --query properties.vaultUri -o tsv)   # ends with "/"

az role assignment create --assignee $(az ad signed-in-user show --query id -o tsv) \
  --role "Key Vault Secrets Officer" --scope $KV_ID
az role assignment create --assignee-object-id $ID_PRINCIPAL --assignee-principal-type ServicePrincipal \
  --role "Key Vault Secrets User" --scope $KV_ID
```

**What:**
- A **user-assigned managed identity**: an identity in Microsoft Entra ID that the containers run
  as. There is no password; Azure issues it tokens.
- A **Key Vault**: holds the database URL, Redis key, storage key and API key.

**Why:** Container Apps fetches secrets from Key Vault *as the identity*, so no secret ever sits
in a config file or the CLI history of the app definition. You assigned yourself *Officer*
(read/write) and the app *User* (read-only): least privilege.

**Why user-assigned rather than system-assigned:** It exists before the apps do, so you can grant it
roles first, and the API and worker can share it. A system-assigned identity is only created along
with the app, which leads to a chicken-and-egg problem for registry pulls.

**Trade-off:** RBAC changes take 1–5 minutes to propagate. If the next `secret set` returns
`Forbidden`, wait and retry. Nothing is wrong.

**Verify:** `az keyvault secret set --vault-name $KV -n probe --value ok && az keyvault secret delete --vault-name $KV -n probe`

---

## 6. Container Registry and images

```bash
az acr create -g $RG -n $ACR --sku Basic
export ACR_ID=$(az acr show -n $ACR --query id -o tsv)
az role assignment create --assignee-object-id $ID_PRINCIPAL --assignee-principal-type ServicePrincipal \
  --role AcrPull --scope $ACR_ID
```

Build the images. **Option A** builds inside Azure, but ACR Tasks are disabled on many
free-trial subscriptions:

```bash
az acr build -r $ACR -t parser-api:v1    apps/api
az acr build -r $ACR -t parser-worker:v1 apps/worker
```

If that fails with `TasksOperationsNotAllowed`, use **option B** and build locally. Your Mac
is ARM, but Container Apps runs `linux/amd64`, so cross-build:

```bash
az acr login -n $ACR
docker buildx build --platform linux/amd64 -t $ACR.azurecr.io/parser-api:v1    --push apps/api
docker buildx build --platform linux/amd64 -t $ACR.azurecr.io/parser-worker:v1 --push apps/worker
```

**Why ACR:** Pulls are authenticated with the managed identity (`AcrPull`), so you never create
registry passwords.

**Trade-off:** Basic tier has no geo-replication or private endpoint, which a single-region app
doesn't need. Building on ARM with emulation is slow (the worker image takes several minutes)
but only happens on the first manual deploy. CI builds natively on amd64.

**Verify:** `az acr repository list -n $ACR -o table` shows `parser-api` and `parser-worker`.

---

## 7. PostgreSQL

```bash
export PG_PASS="$(openssl rand -base64 24 | tr -d '/+=')"
az postgres flexible-server create -g $RG -n $PG -l $LOCATION \
  --tier Burstable --sku-name Standard_B1ms --storage-size 32 --version 16 \
  --admin-user parseradmin --admin-password "$PG_PASS" \
  --public-access 0.0.0.0
az postgres flexible-server firewall-rule create -g $RG -s $PG \
  -n my-laptop --start-ip-address $MYIP --end-ip-address $MYIP

# 001_schema.up.sql runs CREATE EXTENSION pgcrypto. Azure only allows allow-listed extensions.
az postgres flexible-server parameter set -g $RG -s $PG -n azure.extensions -v PGCRYPTO

az postgres flexible-server db create -g $RG -s $PG -d file_platform

export DATABASE_URL="postgres://parseradmin:${PG_PASS}@${PG}.postgres.database.azure.com:5432/file_platform?sslmode=require"
az keyvault secret set --vault-name $KV -n database-url --value "$DATABASE_URL" -o none

./infra/scripts/remote-setup.sh        # migrations + seed + hosted limits (10 MB uploads)
```

**What:** A managed Postgres 16 server, the `file_platform` database, the schema and seed data,
and the hosted upload limits written into `system_settings`.

**Why:**
- `--public-access 0.0.0.0` is Azure's shorthand for "allow connections from Azure services". The
  Container Apps can connect without private networking.
- The firewall rule lets your laptop run the migrations.
- `sslmode=require` is mandatory; Azure enforces TLS.

**Trade-offs:**
- "Allow Azure services" also admits *other customers'* Azure resources, so the strong random
  password plus TLS is what protects you. The stricter option is VNet-integrated (private)
  access, which needs its own delegated subnet and Private DNS zone.
- B1ms is burstable: fine at low traffic, but sustained heavy parsing exhausts its CPU credits.
  The next step up is `Standard_B2s`.

**Verify:** `psql "$DATABASE_URL" -c '\dt'` lists `files`, `ingestion_jobs`, `system_settings` and more.

---

## 8. Redis

```bash
az redis create -g $RG -n $REDIS -l $LOCATION --sku Basic --vm-size c0 --minimum-tls-version 1.2
# ~15–20 min. Start section 9 in a second terminal (re-export the variables there).

export REDIS_HOST=$REDIS.redis.cache.windows.net
export REDIS_KEY=$(az redis list-keys -g $RG -n $REDIS --query primaryKey -o tsv)
az keyvault secret set --vault-name $KV -n redis-key --value "$REDIS_KEY" -o none
```

**Why:** The API pushes jobs onto the Redis list `ingestion_jobs`, and the worker pops them.
The worker's autoscaler (KEDA) watches the same list's length.

**Trade-offs:**
- Basic tier has **no replica and no SLA**. If the node restarts, queued (not yet started) jobs are
  lost. Standard C0 (~$40/month) adds a replica.
- TLS-only means the code needed `REDIS_TLS=true` (done in Phase 0).
- **Retirement notice:** Microsoft is retiring Azure Cache for Redis in favour of **Azure
  Managed Redis**. If `az redis create` is refused for new customers, run:
  ```bash
  az redisenterprise create -g $RG -n $REDIS -l $LOCATION --sku Balanced_B0
  # endpoint: <name>.<region>.redis.azure.net:10000 (TLS). Use port 10000 everywhere 6380 appears below.
  ```

**Verify:** `az redis show -g $RG -n $REDIS --query provisioningState -o tsv` → `Succeeded`.

---

## 9. Blob Storage

```bash
az storage account create -g $RG -n $STORAGE -l $LOCATION \
  --sku Standard_LRS --kind StorageV2 --min-tls-version TLS1_2 \
  --allow-blob-public-access false --https-only true
export STORAGE_KEY=$(az storage account keys list -g $RG -n $STORAGE --query '[0].value' -o tsv)
az storage container create --account-name $STORAGE --account-key "$STORAGE_KEY" -n file-ingestion
az keyvault secret set --vault-name $KV -n storage-key --value "$STORAGE_KEY" -o none
```

(CORS is added in section 12, once the website's URL exists.)

**What:** A storage account with a private container `file-ingestion`. The API also creates the
container itself at startup if it is missing.

**Why direct browser uploads:**
1. The API signs a **SAS URL** that allows create+write on exactly one blob, over HTTPS, for one hour.
2. The browser PUTs the file straight to Blob.
3. File bytes never pass through the API container, so a big upload can't exhaust the API's memory
   or its request timeout.

**Trade-offs:**
- **LRS** keeps 3 copies in one datacenter. ZRS or GRS survive a zone or region outage at higher
  cost; uploaded evidence here can be re-uploaded, so LRS is fine.
- The API signs with the **account key**. The stronger pattern is a *user-delegation SAS* signed
  with the managed identity (role: Storage Blob Data Contributor), which removes the key from
  Key Vault entirely. That's a good follow-up exercise.

**Verify:** `az storage container show --account-name $STORAGE --account-key "$STORAGE_KEY" -n file-ingestion -o table`

---

## 10. OpenSearch VM

```bash
az vm create -g $RG -n vm-search -l $LOCATION \
  --image Ubuntu2204 --size Standard_B2s \
  --vnet-name vnet-parser --subnet snet-search \
  --admin-username azureuser --generate-ssh-keys \
  --public-ip-sku Standard --nsg-rule SSH
az network nsg rule update -g $RG --nsg-name vm-searchNSG -n default-allow-ssh \
  --source-address-prefixes $MYIP/32
export VM_PUBLIC=$(az vm show -d -g $RG -n vm-search --query publicIps -o tsv)
export SEARCH_IP=$(az vm show -d -g $RG -n vm-search --query privateIps -o tsv)
echo "public $VM_PUBLIC  private $SEARCH_IP"

scp infra/opensearch/create-indexes.sh azureuser@$VM_PUBLIC:~
ssh azureuser@$VM_PUBLIC
```

On the VM:

```bash
curl -fsSL https://get.docker.com | sudo sh
echo 'vm.max_map_count=262144' | sudo tee /etc/sysctl.d/99-opensearch.conf && sudo sysctl --system
sudo docker run -d --name opensearch --restart unless-stopped -p 9200:9200 \
  -e discovery.type=single-node -e plugins.security.disabled=true \
  -e OPENSEARCH_JAVA_OPTS="-Xms1g -Xmx1g" \
  -e OPENSEARCH_INITIAL_ADMIN_PASSWORD="$(openssl rand -base64 18)Aa1!" \
  -v osdata:/usr/share/opensearch/data opensearchproject/opensearch:2.13.0
sleep 45 && curl -s localhost:9200 | head -5        # cluster info JSON
OPENSEARCH_URL=http://localhost:9200 bash ~/create-indexes.sh
exit
```

**What:** An Ubuntu VM running the same OpenSearch image as your local stack, with the app's indexes.

**Why the network rules matter:** The NSG (network security group, a stateful firewall) opens
**only** port 22, and only to your IP. Port 9200 is reachable through Azure's default
`AllowVnetInBound` rule, which means from the Container Apps in the same VNet and never from the
internet. That isolation is why running without the security plugin is acceptable, matching local.

**Other notes:**
- `vm.max_map_count` is a kernel limit OpenSearch needs for memory-mapped index files. You can't set
  it in Container Apps, which is another reason for a VM.
- B2s has 4 GB of RAM. A 1 GB heap leaves the rest for the OS file cache, which OpenSearch relies on.

**Trade-offs:**
- You now own patching and backups for this VM. `--restart unless-stopped` brings OpenSearch back
  after a reboot, but there's no snapshot schedule. The index can be rebuilt by re-processing files.
- If B2s isn't offered in your region, try `--size Standard_B2als_v2`.

**Verify (from your laptop):** `curl -m 5 http://$VM_PUBLIC:9200` should **time out**, which
proves it isn't exposed. After section 11, the API's search will prove it is reachable internally.

---

## 11. Container Apps: environment, API and worker

### 11a. Environment

```bash
az containerapp env create -g $RG -n cae-parser -l $LOCATION \
  --logs-destination log-analytics --logs-workspace-id $LOG_ID --logs-workspace-key "$LOG_KEY" \
  --infrastructure-subnet-resource-id $ACA_SUBNET_ID
```

**What:** The shared boundary the apps run in, with its networking inside `snet-aca` and logs going
to the workspace. It's a *workload-profiles* environment using the serverless **Consumption**
profile, so you pay per second of use.

### 11b. Shared settings

```bash
export API_KEY=$(openssl rand -hex 24); echo "API_KEY=$API_KEY   <- save this"
az keyvault secret set --vault-name $KV -n api-key --value "$API_KEY" -o none

KVREF="identityref:$ID_RES"
SECRETS=(
  "database-url=keyvaultref:${KV_URI}secrets/database-url,$KVREF"
  "redis-key=keyvaultref:${KV_URI}secrets/redis-key,$KVREF"
  "storage-key=keyvaultref:${KV_URI}secrets/storage-key,$KVREF"
)
COMMON_ENV=(
  APP_ENV=production QUEUE_NAME=ingestion_jobs
  DATABASE_URL=secretref:database-url
  REDIS_ADDR=$REDIS_HOST:6380 REDIS_PASSWORD=secretref:redis-key REDIS_TLS=true
  STORAGE_BACKEND=azure AZURE_STORAGE_ACCOUNT=$STORAGE AZURE_STORAGE_KEY=secretref:storage-key
  MINIO_BUCKET=file-ingestion
  OPENSEARCH_URL=http://$SEARCH_IP:9200
)
```

**How the secret flow works:**
1. `keyvaultref:` makes Container Apps fetch the value from Key Vault *as `id-parser`*.
2. `secretref:` then exposes it to the process as an environment variable.
3. Rotating a secret means updating it in Key Vault and restarting the revision. Apps re-read
   Key Vault references within about 30 minutes.

### 11c. API

```bash
az containerapp create -g $RG -n ca-parser-api --environment cae-parser \
  --image $ACR.azurecr.io/parser-api:v1 \
  --registry-server $ACR.azurecr.io --registry-identity $ID_RES --user-assigned $ID_RES \
  --ingress external --target-port 8080 \
  --cpu 0.5 --memory 1Gi --min-replicas 0 --max-replicas 2 \
  --secrets "${SECRETS[@]}" "api-key=keyvaultref:${KV_URI}secrets/api-key,$KVREF" \
  --env-vars "${COMMON_ENV[@]}" API_PORT=8080 AUTH_ENABLED=true API_KEY=secretref:api-key \
             ALLOWED_ORIGINS=http://localhost:5188

# Only your IP may call the API. Once any Allow rule exists, every other IP is denied.
az containerapp ingress access-restriction set -g $RG -n ca-parser-api \
  --rule-name my-ip --ip-address $MYIP/32 --action Allow

export API_FQDN=$(az containerapp show -g $RG -n ca-parser-api --query properties.configuration.ingress.fqdn -o tsv)
curl https://$API_FQDN/health
curl -s -o /dev/null -w "%{http_code}\n" https://$API_FQDN/api/v1/jobs                    # 401
curl -s -o /dev/null -w "%{http_code}\n" -H "X-API-Key: $API_KEY" https://$API_FQDN/api/v1/jobs   # 200
```

**Why two locks:** The UI is a public JavaScript bundle, so any key built into it can be read by
anyone who opens the site. The **IP allow-list is the real access control**. The API key only
keeps out scanners and anyone who finds the URL from another network. The data here (parsed
infostealer logs) is sensitive, so neither lock alone would be enough.

**Trade-offs:**
- If your home IP changes, the UI starts returning 403. Re-run the `access-restriction set`
  command with your new IP.
- The proper fix is real user sign-in: Microsoft Entra ID through Container Apps built-in auth,
  or Static Web Apps Standard with a linked backend. Both are good follow-ups.
- `--min-replicas 0` means the first request after an idle period waits a few seconds for a cold
  start. `--min-replicas 1` removes that for about $15/month.

### 11d. Worker

```bash
az containerapp create -g $RG -n ca-parser-worker --environment cae-parser \
  --image $ACR.azurecr.io/parser-worker:v1 \
  --registry-server $ACR.azurecr.io --registry-identity $ID_RES --user-assigned $ID_RES \
  --cpu 1 --memory 2Gi --min-replicas 0 --max-replicas 2 \
  --termination-grace-period 300 \
  --secrets "${SECRETS[@]}" \
  --env-vars "${COMMON_ENV[@]}" WORKER_QUEUE_NAME=ingestion_jobs \
             WORKER_MAX_EXTRACTED_SIZE_MB=200 WORKER_MAX_EXTRACTED_FILES=5000 WORKER_MAX_ARCHIVE_DEPTH=5 \
  --scale-rule-name redis-queue --scale-rule-type redis \
  --scale-rule-metadata "address=$REDIS_HOST:6380" "listName=ingestion_jobs" "listLength=1" "enableTLS=true" \
  --scale-rule-auth "password=redis-key"
```

**What:** A worker with **no ingress**: nothing can call it, it only pulls from the queue. KEDA
checks the Redis list length about every 30 s:
- **0 → 1 replica** when a job appears.
- **Up to 2** when more than one job is waiting.
- **Back to 0** after 300 s with an empty queue.

**Why the grace period:**
1. The worker pops a job *before* processing it, so the list can be empty while it's still busy.
2. When KEDA scales in, Azure sends SIGTERM.
3. The Phase 0 change makes the worker finish the current job, then exit.
4. `--termination-grace-period 300` gives it 5 minutes to do that before SIGKILL.

**Trade-offs:**
- A job longer than ~10 minutes (5 min cooldown + 5 min grace) can still be cut off. For huge
  archives, use the local stack (the hosted cap is 10 MB anyway), or set `--min-replicas 1`
  (~$35/month).
- The textbook Azure alternative is **Container Apps Jobs** (event-driven): one job execution per
  queue item, never killed mid-task. It needs the worker restructured to process one message and
  exit. A good follow-up exercise.

**Verify:**

```bash
az containerapp logs show -g $RG -n ca-parser-api --tail 50 | grep -E "object storage ready|starting API"
az containerapp replica list -g $RG -n ca-parser-worker -o table    # empty: scaled to zero
```

---

## 12. Web UI on Static Web Apps

```bash
az staticwebapp create -g $RG -n swa-parser -l westeurope --sku Free   # SWA runs in a few regions only
export WEB_HOST=$(az staticwebapp show -g $RG -n swa-parser --query defaultHostname -o tsv)

(cd apps/web && npm ci && \
  VITE_API_BASE_URL=https://$API_FQDN/api/v1 VITE_API_KEY=$API_KEY VITE_MAX_UPLOAD_MB=10 npm run build)
npx @azure/static-web-apps-cli deploy apps/web/dist --env production \
  --deployment-token "$(az staticwebapp secrets list -g $RG -n swa-parser --query properties.apiKey -o tsv)"

# Allow the site's origin on both CORS layers:
az containerapp update -g $RG -n ca-parser-api --set-env-vars ALLOWED_ORIGINS=https://$WEB_HOST
az storage cors add --account-name $STORAGE --account-key "$STORAGE_KEY" --services b \
  --methods PUT OPTIONS --origins https://$WEB_HOST \
  --allowed-headers '*' --exposed-headers '*' --max-age 3600
```

**Why:**
- `VITE_*` values are compiled in at **build time**, so a new API URL or key means a rebuild.
- There are **two CORS layers**, because the browser talks to two origins:
  1. the API, for JSON;
  2. Blob storage, for the SAS PUT.
  Each must allow the site's origin explicitly.

**Trade-off:** SWA's region (`westeurope`) is only where its metadata lives; the content itself is
served from a global CDN. The Free tier has no custom-auth or linked-backend features (Standard,
$9/month).

**Verify:** Open `https://$WEB_HOST` in the browser. The dashboard should load. A 403 means your
IP isn't on the API allow-list; a CORS error in DevTools means you need to recheck `ALLOWED_ORIGINS`.

---

## 13. End-to-end check and observing it

```bash
API_BASE_URL=https://$API_FQDN/api/v1 API_KEY=$API_KEY JOB_TIMEOUT_SECONDS=300 ./scripts/e2e-upload.sh
```

The first job takes about 30–60 s longer than locally: KEDA has to notice the queue and start a
worker. Then:
1. Upload a small ZIP through the UI.
2. Watch **Portal → ca-parser-worker → Revisions and replicas**: the count goes 0 → 1, then back
   to 0 about 5 minutes after it finishes.
3. In **Log Analytics → Logs**, run:

```kusto
ContainerAppConsoleLogs_CL
| where ContainerAppName_s == "ca-parser-worker"
| project TimeGenerated, Log_s
| order by TimeGenerated desc

ContainerAppSystemLogs_CL          // KEDA scaling decisions, image pulls, crashes
| where ContainerAppName_s == "ca-parser-worker"
| project TimeGenerated, Reason_s, Log_s
| order by TimeGenerated desc
```

---

## 14. CI/CD with GitHub Actions and OIDC (after the manual deploy works)

```bash
APP_ID=$(az ad app create --display-name gh-parser-deploy --query appId -o tsv)
az ad sp create --id $APP_ID
az ad app federated-credential create --id $APP_ID --parameters '{
  "name": "parser-app-master",
  "issuer": "https://token.actions.githubusercontent.com",
  "subject": "repo:chanchalvdev/parser-app:ref:refs/heads/master",
  "audiences": ["api://AzureADTokenExchange"]
}'
az role assignment create --assignee $APP_ID --role Contributor --scope $(az group show -n $RG --query id -o tsv)

gh secret set AZURE_CLIENT_ID       -b "$APP_ID"
gh secret set AZURE_TENANT_ID       -b "$(az account show --query tenantId -o tsv)"
gh secret set AZURE_SUBSCRIPTION_ID -b "$(az account show --query id -o tsv)"
gh secret set AZURE_STATIC_WEB_APPS_API_TOKEN -b "$(az staticwebapp secrets list -g $RG -n swa-parser --query properties.apiKey -o tsv)"
gh secret set VITE_API_KEY          -b "$API_KEY"
gh variable set AZURE_RG          -b $RG
gh variable set AZURE_ACR         -b $ACR
gh variable set AZURE_API_APP     -b ca-parser-api
gh variable set AZURE_WORKER_APP  -b ca-parser-worker
gh variable set VITE_API_BASE_URL -b https://$API_FQDN/api/v1
```

Then run **Actions → Deploy to Azure → Run workflow** on `master`.

**Why OIDC:**
1. GitHub mints a short-lived token for *this repo and this branch*.
2. Azure trusts it through the federated credential.
3. No client secret exists anywhere, so there's nothing to leak or rotate.

**Trade-offs:**
- `Contributor` on the resource group is broad but contained to this app. Least privilege would be
  `AcrPush` on the registry plus `Container Apps Contributor` on the two apps.
- Images are tagged with the commit SHA, so rolling back means `az containerapp update --image ...:<old-sha>`
  (or re-activating the old revision).

---

## 15. Saving credit and tearing down

```bash
az vm deallocate -g $RG -n vm-search                # stops compute billing (~$1/mo disk remains)
az postgres flexible-server stop -g $RG -n $PG      # Azure auto-starts it after 7 days
# Container Apps already scale to zero. Redis can't be paused: delete and recreate (section 8).

az vm start -g $RG -n vm-search
az postgres flexible-server start -g $RG -n $PG

az group delete -n $RG --yes --no-wait              # removes everything in this guide
```

Key Vault is soft-deleted and keeps its name reserved for 90 days. Free the name with
`az keyvault purge -n $KV`.

Watch spend in **Cost Management → Cost analysis**, filtered to `rg-parser`. Add a budget alert
under **Cost Management → Budgets**.

---

## 16. Fallback

- **Render free tier** (`render.yaml`: Neon, R2, Aiven) still works. It uses the default MinIO/S3
  backend. Set `VITE_API_KEY` on `parser-web` to the generated `API_KEY` of `parser-api`.
- **The local stack** (`make up`) is untouched and remains the place to process multi-GB files.
- Switching storage is one variable (`STORAGE_BACKEND`), so moving the hosted copy between Azure
  and Render needs no code change.

---

## 17. Troubleshooting

| Symptom | Likely cause | Fix |
|---|---|---|
| `remote-setup.sh`: `extension "pgcrypto" is not allow-listed` | Section 7's parameter step was skipped | `az postgres flexible-server parameter set ... -n azure.extensions -v PGCRYPTO` |
| `psql`: connection timed out | Your IP changed | `az postgres flexible-server firewall-rule create -g $RG -s $PG -n my-laptop-2 --start-ip-address <ip> --end-ip-address <ip>` |
| App revision stuck "Provisioning"; system log shows `ImagePull` | `AcrPull` not propagated yet, or the image was built for ARM | Wait 5 min; rebuild with `--platform linux/amd64` |
| System log: secret / Key Vault reference failed | Role assignment not propagated, or a typo in the secret URL | `az keyvault secret show --vault-name $KV -n database-url`; check `${KV_URI}secrets/<name>` |
| API log: `redis unavailable at startup` | Missing `REDIS_TLS=true`, wrong port, or Redis still provisioning | Check env vars and `az redis show ... provisioningState` |
| Browser: 403 from API | IP not on the allow-list | Re-run `access-restriction set` |
| Browser: CORS error on `/uploads/initiate` | `ALLOWED_ORIGINS` doesn't match `https://$WEB_HOST` exactly | `az containerapp update --set-env-vars ALLOWED_ORIGINS=...` |
| Browser: CORS error or 403 on the PUT to `blob.core.windows.net` | Storage CORS is missing the site origin | Section 12 `az storage cors add` |
| Upload 403 `AuthorizationFailure` | Clock skew, or URL edited | The SAS is backdated 5 min; get a fresh URL |
| Job stays `queued` | Worker not scaling | `ContainerAppSystemLogs_CL` for KEDA errors; check the scale-rule metadata and the `redis-key` secret |
| Job fails at indexing | Worker can't reach OpenSearch | `ssh` to the VM: `sudo docker ps`; check `OPENSEARCH_URL` uses the **private** IP |
| Job stuck `running` after a scale-in | Job ran longer than cooldown + grace | Retry it from the UI; consider `--min-replicas 1` |

---

## 18. Where to go next

1. **Infrastructure as code:** rewrite sections 3–12 as one Bicep file and redeploy from scratch
   with a single command. This is the best test of whether you understand every resource.
2. **Identity instead of keys:** Blob via user-delegation SAS, Postgres via Entra ID auth. Key Vault
   then holds nothing at all.
3. **Private networking:** Private Endpoints for Postgres, Redis and Blob, then turn their public
   access off.
4. **Container Apps Jobs** for the worker (section 11d trade-off).
5. **Real user sign-in** with Entra ID, replacing the IP allow-list and shared key.
6. **Application Insights** with OpenTelemetry for request traces across API → queue → worker.
