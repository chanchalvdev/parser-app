import { Navigate, Route, Routes } from 'react-router-dom'
import { DashboardPage } from '@/pages/DashboardPage'
import { UploadPage } from '@/pages/UploadPage'
import { FilesPage } from '@/pages/FilesPage'
import { FileDetailPage } from '@/pages/FileDetailPage'
import { JobsPage } from '@/pages/JobsPage'
import { JobDetailPage } from '@/pages/JobDetailPage'
import { SearchPage } from '@/pages/SearchPage'
import { AdminSettingsPage } from '@/pages/AdminSettingsPage'
import { AuditLogsPage } from '@/pages/AuditLogsPage'

export const AppRouter = () => (
  <Routes>
    <Route path="/" element={<Navigate to="/dashboard" replace />} />
    <Route path="/dashboard" element={<DashboardPage />} />
    <Route path="/upload" element={<UploadPage />} />
    <Route path="/files" element={<FilesPage />} />
    <Route path="/files/:fileId" element={<FileDetailPage />} />
    <Route path="/jobs" element={<JobsPage />} />
    <Route path="/jobs/:jobId" element={<JobDetailPage />} />
    <Route path="/search" element={<SearchPage />} />
    <Route path="/admin/settings" element={<AdminSettingsPage />} />
    <Route path="/audit-logs" element={<AuditLogsPage />} />
    <Route path="*" element={<Navigate to="/dashboard" replace />} />
  </Routes>
)
