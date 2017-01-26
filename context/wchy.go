package context

import "github.com/WeCanHearYou/wchy-api/services"

// WchySettings is an application-wide settings
type WchySettings struct {
	BuildTime string
}

// WchyContext is an application-wide context
type WchyContext struct {
	Health   services.HealthCheckService
	Tenant   services.TenantService
	Settings WchySettings
}
