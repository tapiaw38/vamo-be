package web

import (
	"log/slog"

	"github.com/gin-gonic/gin"
	"github.com/jackc/pgx/v5/pgxpool"
	adminrepo "github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/admin"
	bookingrepo "github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/booking"
	paymentrepo "github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/bookingpayment"
	companyrepo "github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/company"
	driverrepo "github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/driver"
	locationrepo "github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/location"
	pricingrepo "github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/pricing"
	ratingrepo "github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/rating"
	riderequest "github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/riderequest"
	triprepo "github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/trip"
	userrepo "github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/user"
	vehiclerepo "github.com/tapiaw38/vamo/backend/internal/adapters/datasources/repositories/vehicle"
	adminh "github.com/tapiaw38/vamo/backend/internal/adapters/web/handlers/admin"
	bookingh "github.com/tapiaw38/vamo/backend/internal/adapters/web/handlers/bookings"
	companyh "github.com/tapiaw38/vamo/backend/internal/adapters/web/handlers/companies"
	driverhandlers "github.com/tapiaw38/vamo/backend/internal/adapters/web/handlers/drivers"
	geocodehandlers "github.com/tapiaw38/vamo/backend/internal/adapters/web/handlers/geocode"
	"github.com/tapiaw38/vamo/backend/internal/adapters/web/handlers/health"
	locationhandlers "github.com/tapiaw38/vamo/backend/internal/adapters/web/handlers/locations"
	paymenth "github.com/tapiaw38/vamo/backend/internal/adapters/web/handlers/payments"
	pricinghandlers "github.com/tapiaw38/vamo/backend/internal/adapters/web/handlers/pricing"
	ratinghandlers "github.com/tapiaw38/vamo/backend/internal/adapters/web/handlers/ratings"
	registrationh "github.com/tapiaw38/vamo/backend/internal/adapters/web/handlers/registration"
	ridehandlers "github.com/tapiaw38/vamo/backend/internal/adapters/web/handlers/riderequests"
	triph "github.com/tapiaw38/vamo/backend/internal/adapters/web/handlers/trips"
	userhandlers "github.com/tapiaw38/vamo/backend/internal/adapters/web/handlers/users"
	vehiclehandlers "github.com/tapiaw38/vamo/backend/internal/adapters/web/handlers/vehicles"
	"github.com/tapiaw38/vamo/backend/internal/adapters/web/integrations/authapi"
	paymentapi "github.com/tapiaw38/vamo/backend/internal/adapters/web/integrations/payments"
	"github.com/tapiaw38/vamo/backend/internal/platform/auth"
	"github.com/tapiaw38/vamo/backend/internal/platform/config"
	apperrors "github.com/tapiaw38/vamo/backend/internal/platform/errors"
	platformlogging "github.com/tapiaw38/vamo/backend/internal/platform/logging"
	"github.com/tapiaw38/vamo/backend/internal/platform/storage"
	adminuc "github.com/tapiaw38/vamo/backend/internal/usecases/admin"
	bookinguc "github.com/tapiaw38/vamo/backend/internal/usecases/bookings"
	companyuc "github.com/tapiaw38/vamo/backend/internal/usecases/companies"
	driveruc "github.com/tapiaw38/vamo/backend/internal/usecases/drivers"
	geocodeuc "github.com/tapiaw38/vamo/backend/internal/usecases/geocode"
	locationuc "github.com/tapiaw38/vamo/backend/internal/usecases/locations"
	paymentuc "github.com/tapiaw38/vamo/backend/internal/usecases/payments"
	pricinguc "github.com/tapiaw38/vamo/backend/internal/usecases/pricing"
	profilebootstrap "github.com/tapiaw38/vamo/backend/internal/usecases/profilebootstrap"
	ratinguc "github.com/tapiaw38/vamo/backend/internal/usecases/ratings"
	registrationuc "github.com/tapiaw38/vamo/backend/internal/usecases/registration"
	rideuc "github.com/tapiaw38/vamo/backend/internal/usecases/riderequests"
	routinguc "github.com/tapiaw38/vamo/backend/internal/usecases/routing"
	tripuc "github.com/tapiaw38/vamo/backend/internal/usecases/trips"
	useruc "github.com/tapiaw38/vamo/backend/internal/usecases/users"
	vehicleuc "github.com/tapiaw38/vamo/backend/internal/usecases/vehicles"
	"time"
)

func RegisterRoutes(router *gin.Engine, db *pgxpool.Pool, logger *slog.Logger, authClient authapi.Client, jwtSecret string, paymentsEnabled bool, paymentsURL, paymentsKey, paymentsSecret, serviceEmail, servicePassword string, s3Config config.S3Config) {
	router.Use(gin.Recovery(), apperrors.Middleware(), platformlogging.Middleware(logger))
	router.GET("/health", health.NewGetHandler())
	router.GET("/ready", health.NewReadyHandler(db))
	router.GET("/api/v1/openapi.yaml", func(c *gin.Context) { c.File("docs/openapi.yaml") })
	router.POST("/api/v1/auth/register", registrationh.NewRegisterHandler(registrationuc.Usecase{Auth: authClient, Profile: userrepo.NewRepository(db), ServiceEmail: serviceEmail, ServiceSecret: servicePassword, Logger: logger}))

	locations := locationrepo.NewRepository(db)
	router.GET("/api/v1/provinces", locationhandlers.NewListHandler(locationuc.NewListUsecase(locations)))
	router.GET("/api/v1/provinces/:id/localities", locationhandlers.NewListLocalitiesHandler(locationuc.NewListLocalitiesUsecase(locations)))
	router.GET("/api/v1/localities/search", locationhandlers.NewSearchHandler(locationuc.NewSearchUsecase(locations)))
	router.GET("/api/v1/localities/nearest", locationhandlers.NewNearestHandler(locationuc.NewNearestUsecase(locations)))
	geocodeService := geocodeuc.NewService(nil, "")
	routingService := routinguc.NewService(nil, "")
	geocodeRoutes := router.Group("/api/v1/geocode", auth.Middleware(jwtSecret))
	geocodeRoutes.GET("/reverse", geocodehandlers.NewReverse(geocodeService))
	geocodeRoutes.GET("/search", geocodehandlers.NewSearch(geocodeService))
	geocodeRoutes.GET("/route", geocodehandlers.NewRoute(routingService))

	profiles := useruc.NewGetUsecase(userrepo.NewRepository(db), authClient)
	updates := useruc.NewUpdateUsecase(userrepo.NewRepository(db))
	protected := router.Group("/api/v1/users", auth.Middleware(jwtSecret))
	protected.GET("/me", userhandlers.NewGetHandler(profiles))
	protected.PATCH("/me", userhandlers.NewUpdateHandler(updates))
	protected.GET("/search", userhandlers.NewSearchHandler(useruc.NewSearchUsecase(userrepo.NewRepository(db))))

	drivers := driverrepo.NewRepository(db)
	vehicles := vehiclerepo.NewRepository(db)
	images := storage.NewS3ImageStorage(s3Config)
	driverUsecase := driveruc.NewSaveUsecase(drivers)
	protectedDrivers := router.Group("/api/v1/drivers", auth.Middleware(jwtSecret))
	protectedDrivers.GET("/me", driverhandlers.NewGetHandler(driveruc.NewGetUsecase(drivers)))
	protectedDrivers.POST("/profile", auth.RequireConductor(), driverhandlers.NewSaveHandler(driverUsecase))
	protectedDrivers.PATCH("/me", driverhandlers.NewSaveHandler(driverUsecase))
	protectedDrivers.PATCH("/me/availability", auth.RequireConductor(), driverhandlers.NewAvailabilityHandler(driveruc.NewAvailabilityUsecase(drivers)))

	protectedVehicles := router.Group("/api/v1/vehicles", auth.Middleware(jwtSecret))
	protectedVehicles.POST("", auth.RequireConductor(), vehiclehandlers.NewCreateHandler(vehicleuc.NewCreateUsecase(drivers, vehicles, images)))
	protectedVehicles.GET("", vehiclehandlers.NewListHandler(vehicleuc.NewListUsecase(drivers, vehicles, images)))
	protectedVehicles.GET("/:id", vehiclehandlers.NewGetHandler(vehicleuc.NewGetUsecase(drivers, vehicles, images)))
	protectedVehicles.PATCH("/:id", vehiclehandlers.NewUpdateHandler(vehicleuc.NewUpdateUsecase(drivers, vehicles, images)))

	companies := companyrepo.NewRepository(db)
	trips := triprepo.NewRepository(db)
	pricing := pricinguc.NewUsecases(pricingrepo.NewRepository(db))
	profileBootstrap := profilebootstrap.Service{Profile: userrepo.NewRepository(db), Auth: authClient, Logger: logger}
	protectedTrips := router.Group("/api/v1/trips", auth.Middleware(jwtSecret))
	protectedTrips.POST("", auth.RequireConductor(), triph.NewCreateHandler(tripuc.NewCreateUsecase(drivers, vehicles, trips, companies)))
	protectedTrips.PATCH("/:id", triph.NewUpdateHandler(tripuc.NewUpdateUsecase(drivers, vehicles, trips)))
	protectedTrips.POST("/:id/cancel", triph.NewCancelHandler(tripuc.NewCancelUsecase(drivers, trips)))
	protectedTrips.GET("/mine", triph.NewMineHandler(tripuc.NewMineUsecase(drivers, trips)))
	publicTrips := router.Group("/api/v1/trips")
	publicTrips.GET("", triph.NewSearchHandler(tripuc.NewSearchUsecase(trips, images)))
	publicTrips.GET("/:id", triph.NewGetHandler(tripuc.NewGetUsecase(trips, images)))

	rideRequests := rideuc.Usecases{Repo: riderequest.NewRepository(db), Drivers: drivers, Pricing: pricing, Profile: profileBootstrap, Logger: logger}
	protectedRideRequests := router.Group("/api/v1/ride-requests", auth.Middleware(jwtSecret))
	protectedRideRequests.GET("/quote", pricinghandlers.NewQuote(pricing))
	protectedRideRequests.POST("", ridehandlers.NewCreate(rideRequests))
	protectedRideRequests.GET("/active", ridehandlers.NewActive(rideRequests))
	protectedRideRequests.GET("/recent", ridehandlers.NewRecent(rideRequests))
	protectedRideRequests.GET("", auth.RequireConductor(), ridehandlers.NewList(rideRequests))
	protectedRideRequests.POST("/:id/accept", auth.RequireConductor(), ridehandlers.NewAccept(rideRequests))
	protectedRideRequests.POST("/:id/cancel", ridehandlers.NewCancel(rideRequests))

	bookings := bookingrepo.NewRepository(db)
	ratings := ratinguc.Usecases{Repo: ratingrepo.NewRepository(db)}
	protectedBookings := router.Group("/api/v1", auth.Middleware(jwtSecret))
	protectedBookings.POST("/trips/:id/bookings", bookingh.NewCreateHandler(bookinguc.NewCreateUsecase(bookings, profileBootstrap)))
	protectedBookings.GET("/bookings/mine", bookingh.NewListHandler(bookinguc.NewListUsecase(bookings)))
	protectedBookings.GET("/bookings/received", bookingh.NewReceivedHandler(bookinguc.NewReceivedUsecase(bookings)))
	protectedBookings.GET("/bookings/:id", bookingh.NewGetHandler(bookinguc.NewGetUsecase(bookings)))
	protectedBookings.POST("/bookings/:id/cancel", bookingh.NewCancelHandler(bookinguc.NewCancelUsecase(bookings)))
	protectedBookings.POST("/bookings/:id/confirm", bookingh.NewConfirmHandler(bookinguc.NewConfirmUsecase(bookings)))
	protectedBookings.POST("/bookings/:id/reject", bookingh.NewRejectHandler(bookinguc.NewRejectUsecase(bookings)))
	protectedBookings.POST("/bookings/:id/rating", ratinghandlers.NewCreate(ratings))
	publicDrivers := router.Group("/api/v1/drivers")
	publicDrivers.GET("/:id/rating", ratinghandlers.NewDriverStats(ratings))

	companyUsecases := companyuc.Usecases{Companies: companies, VehicleRepo: vehicles}
	companyRoutes := router.Group("/api/v1/companies", auth.Middleware(jwtSecret))
	companyRoutes.POST("", companyh.NewCreate(companyUsecases))
	companyRoutes.GET("", companyh.NewList(companyUsecases))
	companyRoutes.GET("/:id", companyh.NewGet(companyUsecases))
	companyRoutes.PATCH("/:id", companyh.NewUpdate(companyUsecases))
	companyRoutes.GET("/:id/members", companyh.NewMembers(companyUsecases))
	companyRoutes.POST("/:id/members", companyh.NewAddMember(companyUsecases))
	companyRoutes.DELETE("/:id/members/:memberID", companyh.NewRemoveMember(companyUsecases))
	companyRoutes.GET("/:id/vehicles", companyh.NewVehicles(companyUsecases))
	companyRoutes.POST("/:id/vehicles", companyh.NewAddVehicle(companyUsecases))

	adminUsecases := adminuc.Usecases{Repo: adminrepo.NewRepository(db)}
	adminRoutes := router.Group("/api/v1/admin", auth.Middleware(jwtSecret), auth.AdminMiddleware())
	adminRoutes.GET("/dashboard", adminh.NewDashboard(adminUsecases))
	adminRoutes.GET("/users", adminh.NewList(adminUsecases, "users"))
	adminRoutes.GET("/drivers", adminh.NewList(adminUsecases, "drivers"))
	adminRoutes.GET("/companies", adminh.NewList(adminUsecases, "companies"))
	adminRoutes.GET("/trips", adminh.NewList(adminUsecases, "trips"))
	adminRoutes.GET("/bookings", adminh.NewList(adminUsecases, "bookings"))
	adminRoutes.POST("/verify/:kind/:id", adminh.NewVerify(adminUsecases))
	adminRoutes.GET("/pricing", pricinghandlers.NewGet(pricing))
	adminRoutes.PUT("/pricing", pricinghandlers.NewUpdate(pricing))

	paymentUsecases := paymentuc.Usecases{Repo: paymentrepo.NewRepository(db), Client: paymentapi.NewClient(paymentsURL, paymentsKey, 10*time.Second), Enabled: paymentsEnabled}
	protectedPayments := router.Group("/api/v1", auth.Middleware(jwtSecret))
	protectedPayments.POST("/bookings/:id/payment", paymenth.NewStart(paymentUsecases))
	// Payments webhook uses its own HMAC, not a user JWT. Replace client/config at composition root when enabling.
	router.POST("/api/v1/internal/payments/webhook", paymenth.NewWebhook(paymentUsecases, paymentsSecret))
}
