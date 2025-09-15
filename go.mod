module gateway-cd

go 1.21

require (
	github.com/gin-gonic/gin v1.9.1
	github.com/go-logr/logr v1.3.0
	github.com/prometheus/client_golang v1.17.0
	k8s.io/api v0.28.4
	k8s.io/apimachinery v0.28.4
	k8s.io/client-go v0.28.4
	sigs.k8s.io/controller-runtime v0.16.3
	sigs.k8s.io/gateway-api v1.0.0
	gorm.io/driver/postgres v1.5.4
	gorm.io/gorm v1.25.5
)