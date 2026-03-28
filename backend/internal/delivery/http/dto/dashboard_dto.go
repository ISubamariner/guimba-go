package dto

type DashboardStatsResponse struct {
	TotalTenants    int `json:"total_tenants"`
	TotalProperties int `json:"total_properties"`
	ActiveDebts     int `json:"active_debts"`
	OverdueDebts    int `json:"overdue_debts"`
}

type RecentActivityResponse struct {
	Action      string `json:"action"`
	Description string `json:"description"`
	Timestamp   string `json:"timestamp"`
}

type RecentActivitiesResponse struct {
	Data []RecentActivityResponse `json:"data"`
}
