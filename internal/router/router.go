package router

import (
	"net/http"

	"github.com/go-chi/chi/v5"
	chimiddleware "github.com/go-chi/chi/v5/middleware"

	// _ "github.com/knnedy/projectflow/docs"
	"github.com/knnedy/projectflow/internal/handler"
	"github.com/knnedy/projectflow/internal/middleware"
	"github.com/knnedy/projectflow/internal/repository"
	"github.com/knnedy/projectflow/internal/token"
)

func New(
	db *repository.DB,
	tokens *token.TokenManager,
	auth *handler.AuthHandler,
	user *handler.UserHandler,
	org *handler.OrgHandler,
	member *handler.MemberHandler,
	project *handler.ProjectHandler,
	issue *handler.IssueHandler,
	comment *handler.CommentHandler,
	activity *handler.ActivityHandler,
) http.Handler {
	r := chi.NewRouter()

	// global middleware
	r.Use(chimiddleware.RequestID)
	r.Use(chimiddleware.RealIP)
	r.Use(chimiddleware.Recoverer)
	r.Use(middleware.Logger)

	// swagger docs
	// r.Get("/docs/*", httpSwagger.Handler(
	// 	httpSwagger.URL("/docs/doc.json"),
	// ))

	// init middleware
	authMw := middleware.NewAuthMiddleware(tokens)
	orgMw := middleware.NewOrgMiddleware(db.Queries())

	r.Route("/api/v1", func(r chi.Router) {

		// auth — public
		r.Route("/auth", func(r chi.Router) {
			r.Post("/register", auth.Register)
			r.Post("/login", auth.Login)
			r.Post("/refresh", auth.RefreshAccessToken)
			r.Post("/logout", auth.Logout)
		})

		// invitation accept — authenticated but no org context needed
		r.Group(func(r chi.Router) {
			r.Use(authMw.Authenticate)
			r.Post("/invitations/accept", org.AcceptInvitation)
		})

		// authenticated routes
		r.Group(func(r chi.Router) {
			r.Use(authMw.Authenticate)

			// users
			r.Route("/users", func(r chi.Router) {
				r.Get("/me", user.GetMe)
				r.Patch("/me", user.UpdateMe)
				r.Patch("/me/password", user.UpdatePassword)
				r.Delete("/me", user.DeleteMe)
			})

			// organisations
			r.Route("/organisations", func(r chi.Router) {
				r.Post("/", org.Create)
				r.Get("/", org.List)

				r.Route("/{orgID}", func(r chi.Router) {
					r.Use(orgMw.ResolveOrg)

					// org CRUD
					r.Get("/", org.GetByID)
					r.With(middleware.RequireRole(repository.MemberRoleOWNER, repository.MemberRoleADMIN)).Patch("/", org.Update)
					r.With(middleware.RequireRole(repository.MemberRoleOWNER)).Delete("/", org.Delete)

					// members
					r.Route("/members", func(r chi.Router) {
						r.Get("/", member.List)
						r.Delete("/me", member.LeaveOrg)
						r.With(middleware.RequireRole(repository.MemberRoleOWNER, repository.MemberRoleADMIN)).Post("/invitations", org.InviteMember)
						r.With(middleware.RequireRole(repository.MemberRoleOWNER, repository.MemberRoleADMIN)).Patch("/{memberID}", member.UpdateMemberRole)
						r.With(middleware.RequireRole(repository.MemberRoleOWNER, repository.MemberRoleADMIN)).Delete("/{memberID}", member.Delete)
					})

					// org activity
					r.Get("/activity", activity.ListByOrg)

					// projects
					r.Route("/projects", func(r chi.Router) {
						r.With(middleware.RequireRole(repository.MemberRoleOWNER, repository.MemberRoleADMIN)).Post("/", project.Create)
						r.Get("/", project.List)

						r.Route("/{projectID}", func(r chi.Router) {
							r.Get("/", project.GetByID)
							r.With(middleware.RequireRole(repository.MemberRoleOWNER, repository.MemberRoleADMIN)).Patch("/", project.Update)
							r.With(middleware.RequireRole(repository.MemberRoleOWNER, repository.MemberRoleADMIN)).Delete("/", project.Delete)

							// project activity
							r.Get("/activity", activity.ListByProject)

							// issues
							r.Route("/issues", func(r chi.Router) {
								r.With(middleware.RequireRole(repository.MemberRoleOWNER, repository.MemberRoleADMIN)).Post("/", issue.Create)
								r.Get("/", issue.List)

								r.Route("/{issueID}", func(r chi.Router) {
									r.Get("/", issue.GetByID)
									r.Patch("/", issue.UpdateDetails)
									r.Patch("/status", issue.UpdateStatus)
									r.With(middleware.RequireRole(repository.MemberRoleOWNER, repository.MemberRoleADMIN)).Delete("/", issue.Delete)

									// comments
									r.Route("/comments", func(r chi.Router) {
										r.Post("/", comment.Create)
										r.Get("/", comment.List)
										r.Patch("/{commentID}", comment.Update)
										r.Delete("/{commentID}", comment.Delete)
									})
								})
							})
						})
					})
				})
			})

			// entity activity — accessible by authenticated users
			r.Get("/activity/{entityID}", activity.ListByEntity)
		})
	})

	return r
}
