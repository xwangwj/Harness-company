export default function Home() {
  return (
    <div className="flex min-h-screen items-center justify-center">
      <div className="text-center space-y-4">
        <h1 className="text-4xl font-bold text-slate-900">Harness Organization</h1>
        <p className="text-lg text-slate-500">Self-evolving organizational management platform</p>
        <div className="flex gap-4 justify-center pt-4">
          <a href="/login" className="px-6 py-2 bg-primary-600 text-white rounded-lg hover:bg-primary-700 transition">
            Sign In
          </a>
          <a href="/register" className="px-6 py-2 border border-slate-300 rounded-lg hover:bg-slate-100 transition">
            Register
          </a>
        </div>
      </div>
    </div>
  )
}
