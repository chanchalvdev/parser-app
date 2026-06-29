export const Section = ({ title, children }: { title?: string; children: React.ReactNode }) => {
  return (
    <section className="panel p-4">
      {title ? <h2 className="panel-title mb-3 text-white">{title}</h2> : null}
      {children}
    </section>
  )
}
