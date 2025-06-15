package router

import (
	"fmt"
	"github.com/carinfinin/keeper/internal/config"
	"github.com/carinfinin/keeper/internal/store/models"
	"github.com/go-chi/chi"
	"net/http"
)

type ServiceInterface interface {
}

type Router struct {
	Handler *chi.Mux
	service ServiceInterface
	Config  *config.Config
}

func New(cfg *config.Config, service ServiceInterface) *Router {
	return &Router{
		Handler: chi.NewRouter(),
		service: service,
		Config:  cfg,
	}
}

func (r *Router) Configure() {

	r.Handler.Route("/api", func(cr chi.Router) {
		cr.Post("/register", r.Register)
		cr.Post("/login", r.Login)

		//cr.With(r.AuthMiddleware).Post("/orders", r.OrderSave)
	})
}

func (r *Router) Register(writer http.ResponseWriter, request *http.Request) {

	t, err := Generate(&models.User{}, "access", r.Config)
	if err != nil {
		fmt.Println(err)
	}
	fmt.Println(t)

	writer.Write([]byte(t))
}

func (r *Router) Login(writer http.ResponseWriter, request *http.Request) {
	_, err := Decode("eyJhbGciOiJSUzI1NiIsImtpZCI6InYxIiwidHlwIjoiSldUIn0.eyJhdWQiOiJleGFtcGxlLWFwcCIsImV4cCI6MTc0OTgxMDk2NSwiaWF0IjoxNzQ5ODEwMDY1LCJpc3MiOiJodHRwczovL2F1dGguZXhhbXBsZS5jb20iLCJuYmYiOjE3NDk4MTAwNjUsInByZWZlcnJlZF91c2VybmFtZSI6IiIsInNjb3BlIjoib3BlbmlkIHByb2ZpbGUgZW1haWwiLCJzdWIiOjAsInR5cGUiOiJhY2Nlc3MifQ.dI2D5UV_q-Hb-ZTAs5hKWA_sdOA5EbHGg8YJ2kW0hiUJJCdLkMk3nSIK_tIWBzFiH_apz8AjEG7wBbwP5-5yBrGtOosKnDpwezmVUSB8H7I_ZaoUNCGZGPza-siOVhJgOJmSLFCsXmK7pXZIdN5_eyhNGoLL2Ib0G_uv3gTSV3zuZm7y8m9lT7vfoNa97NDIzQfePTpvzzkM6I8kJ2LxAlHATHJxVyFQwT2jA6FFEgLFFs2dgqoVfjefMzZacgfSXoc4biVwzFs0I1U9LsD7J9WWu064SRQp1TA0ce68K9SyAfEKWZiVnC2FwiULmFeaFdjuPdI_AX8qxVaa5Gia_G7GRRhVIYwZL8AVO7hJGbyo3K2W0EKNnwVH_CORoM9PYWiNT7j7IIonqH-lqKtqcEMrHHdusalFAZgiuDuWfMKp0Rdt5iR0A4hY6YrTVLVgPPKdc5q7JrF3mgREcDEtvT2XW6mXHWhxvv69qoIVfObhN6J6CzdiHfKDabvasegV8OWpyhC4OV9nNoOvTCZmZ-LLkhawRoIIQYSS08eqJTl0U8gYYZewllESiBhvmGDOVs5cR02gziBFFhz5Pgnu3sWyEDZmuB6SKhi4_DPhtFKmuQ52q5lOh0IFv_qoXs1ExJh4tc3Jht8iCADgi4NrtX3lu2FefyQx3I2F1cuwHEQ",
		r.Config)
	if err != nil {
		fmt.Println(err)
	}
}
