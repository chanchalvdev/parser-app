#!/usr/bin/env bash
set -euo pipefail

cmd="${1:-help}"
case "$cmd" in
  review)
    shift
    bash harness/scripts/review.sh "$@"
    ;;
  design)
    shift
    bash harness/scripts/design.sh "$@"
    ;;
  plan)
    shift
    bash harness/scripts/plan.sh "$@"
    ;;
  agree)
    shift
    bash harness/scripts/agree.sh
    ;;
  test)
    shift
    bash harness/scripts/test-run.sh
    ;;
  validate)
    shift
    bash harness/scripts/validate.sh
    ;;
  pipeline)
    shift
    bash harness/scripts/pipeline.sh "$@"
    ;;
  harness)
    shift
    bash harness/scripts/pipeline.sh "$@"
    ;;
  self_improve|self-improve)
    shift
    bash harness/scripts/self-improve.sh "$@"
    ;;
  release)
    shift
    bash harness/scripts/release.sh
    ;;
  help|*)
    cat <<'USAGE'
Usage: bash harness/run.sh <command> [args]

Commands:
  review <scope>
  design <agent>
  plan <scope>
  agree
  test
  validate
  pipeline [--scope "<scope>"] [--agent "<agent>"] [--mode <light|full>] [--self-improve "<note>"]
  harness [--scope "<scope>"] [--agent "<agent>"] [--mode <light|full>] [--self-improve "<note>"]
  self_improve|self-improve "<note>"
  release
USAGE
    ;;
esac
