package routes

import (
	"gobackend/controller"
	"gobackend/middleware"

	"github.com/gofiber/fiber/v2"
)



func RegisterAPIKeyRoutes(app *fiber.App){
		// handling normal routings

	auth:= app.Group("/api/auth")
	auth.Get("/getUser",middleware.FetchUser(), controller.GetUser())
	
	// handling normal routings with auth middleware
	project := app.Group("/api/project", middleware.FetchUser())
	
	project.Post("/createProjet", controller.CreateProject())
	project.Put("/updateProject/:projectid", controller.UpdateProject())
	project.Get("/readProject", controller.ReadProject())
	project.Get("/readProjectWithId/:projectid",controller.FindOneViaPID())
	project.Delete("/deleteProject/:projectid",controller.DeleteProject())
	// Blog-section
	blog := app.Group("/api/blog", middleware.FetchUser())

	blog.Post("/createBlog",controller.CreateBlog())
	blog.Get("/readBlog", controller.ReadBlog())
	blog.Get("/readBlog/:blogid", controller.ReadBlogWIthId())
	blog.Put("/updateBlog/:blogid",controller.UpdateBlogWithBlogId())
	blog.Delete("/deleteBlog/:blogid", controller.DeleteBlog())
	
	
	// Link section
	// this section can be made as your second brain
	link := app.Group("/api/link", middleware.FetchUser())
	link.Post("/createLink",  controller.CreateLink())
	link.Get("/readLink",  controller.ReadLink())
	link.Get("/readLink/:linkid",  controller.ReadLinkWithLinkId())
	link.Put("/updateLink/:linkid", controller.UpdateLinkWithLinkId())
	link.Delete("/deleteLink/:linkid", controller.DeleteLinkWithLinkId())
	link.Delete("/deleteAllLink", controller.DeleteAllLinks())
}