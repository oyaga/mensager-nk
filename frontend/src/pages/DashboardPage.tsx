import { useQuery } from '@tanstack/react-query'
import { api } from '../lib/api'
import { MessageSquare, Users, CheckCircle, Clock } from 'lucide-react'

export default function DashboardPage() {
  const { data: stats } = useQuery({
    queryKey: ['stats'],
    queryFn: async () => {
      const response = await api.get('/admin/stats')
      return response.data
    },
  })

  const statCards = [
    {
      name: 'Total Conversations',
      value: stats?.total_conversations || 0,
      icon: MessageSquare,
      color: 'bg-blue-500',
    },
    {
      name: 'Open Conversations',
      value: stats?.open_conversations || 0,
      icon: Clock,
      color: 'bg-yellow-500',
    },
    {
      name: 'Total Contacts',
      value: stats?.total_contacts || 0,
      icon: Users,
      color: 'bg-green-500',
    },
    {
      name: 'Total Messages',
      value: stats?.total_messages || 0,
      icon: CheckCircle,
      color: 'bg-purple-500',
    },
  ]

  return (
    <div className="p-8">
      <div className="mb-8">
        <h1 className="text-3xl font-bold text-gray-900">Dashboard</h1>
        <p className="text-gray-600 mt-2">Welcome back! Here's what's happening today.</p>
      </div>

      {/* Stats Grid */}
      <div className="grid grid-cols-1 md:grid-cols-2 lg:grid-cols-4 gap-6 mb-8">
        {statCards.map((stat) => (
          <div key={stat.name} className="card">
            <div className="flex items-center justify-between">
              <div>
                <p className="text-sm text-gray-600 mb-1">{stat.name}</p>
                <p className="text-3xl font-bold text-gray-900">{stat.value}</p>
              </div>
              <div className={`w-12 h-12 ${stat.color} rounded-lg flex items-center justify-center`}>
                <stat.icon className="w-6 h-6 text-white" />
              </div>
            </div>
          </div>
        ))}
      </div>

      {/* Recent Activity */}
      <div className="card">
        <h2 className="text-xl font-bold text-gray-900 mb-4">Recent Activity</h2>
        <div className="text-center py-12 text-gray-500">
          <MessageSquare className="w-12 h-12 mx-auto mb-4 text-gray-400" />
          <p>No recent activity</p>
        </div>
      </div>
    </div>
  )
}
