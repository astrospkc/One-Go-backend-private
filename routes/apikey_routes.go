package routes

import (
	"gobackend/controller"
	"gobackend/middleware"

	"github.com/gofiber/fiber/v2"
)



func RegisterAPIKeyRoutes(app *fiber.App){
		// handling normal routings

	auth:= app.Group("/api/auth", middleware.ValidateAPIKey())
	auth.Get("/getUser", controller.GetUser())

	collection:=app.Group("/api/collection", middleware.ValidateAPIKey())
	collection.Post("/", controller.CreateCollection())
	collection.Get("/", controller.GetAllCollection())
	
	// handling normal routings with auth middleware
	project := app.Group("/api/project", middleware.ValidateAPIKey())
	
	project.Post("/:col_id", controller.CreateProject())
	project.Put("/:projectid", controller.UpdateProject())
	project.Get("/", controller.GetAllProject())
	project.Get("/collectionProject/:col_id", controller.GetAllProjectOfCollectionId())
	project.Get("/readProject/:projectid",controller.GetAllProjectOfCollectionId())
	project.Delete("/deleteProject/:projectid",controller.DeleteProject())
	project.Delete("/deleteAllProject/:col_id", controller.DeleteAllProject())

	// Blog-section
	blog := app.Group("/api/blog", middleware.ValidateAPIKey())

	blog.Post("/createBlog/:col_id",controller.CreateBlog())
	blog.Get("/readAllBlog", controller.ReadBlog())
	blog.Get("/readAllBlogWithCol_id/:col_id", controller.ReadBlogWithCollectionId())
	blog.Get("/readOneBlog/:blogid", controller.ReadBlogWIthId())
	blog.Put("/updateBlog/:blogid",controller.UpdateBlogWithBlogId())
	blog.Delete("/deleteBlog/:blogid", controller.DeleteBlog())
	blog.Delete("/deleteAllBlog/:col_id", controller.DeleteAllBlog())
	
	
	// Link section
	// this section can be made as your second brain
	link := app.Group("/api/link", middleware.ValidateAPIKey())
	link.Post("/createLink",  controller.CreateLink())
	link.Get("/readLink",  controller.ReadLink())
	link.Get("/readLink/:linkid",  controller.ReadLinkWithLinkId())
	link.Put("/updateLink/:linkid", controller.UpdateLinkWithLinkId())
	link.Delete("/deleteLink/:linkid", controller.DeleteLinkWithLinkId())
	link.Delete("/deleteAllLink", controller.DeleteAllLinks())

	media := app.Group("/api/media", middleware.ValidateAPIKey())
	media.Post("/postmedia/:col_id", controller.PostMedia())
	media.Get("/getALlMediaFiles/:col_id", controller.GetAllMediaFiles())
}