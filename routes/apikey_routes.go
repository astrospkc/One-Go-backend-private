package routes

import (
	"gobackend/controller"
	"gobackend/middleware"

	"github.com/gofiber/fiber/v2"
)



func RegisterAPIKeyRoutes(app *fiber.App){
		// handling normal routings

	auth:= app.Group("/api/v1/auth", middleware.ValidateAPIKey())
	auth.Get("/getUser", controller.GetUser())

	collection:=app.Group("/api/v1/collection", middleware.ValidateAPIKey())
	collection.Get("with-projects", controller.FetchAllCollectionWithProjects())
	collection.Post("/", middleware.IsSubscribed(), controller.CreateCollection(),)
	collection.Get("/", controller.GetAllCollection())
	collection.Delete("/deleteAllCollection", middleware.IsSubscribed(), controller.DeleteAllCollection())
	collection.Delete("/:id", middleware.IsSubscribed(), controller.DeleteCollection())
	collection.Put("/:id", middleware.IsSubscribed(), controller.UpdateCollection())
	collection.Get("/:id", controller.GetCollection())
	
	// handling normal routings with auth middleware
	project := app.Group("/api/v1/project", middleware.ValidateAPIKey())
	
	project.Post("/:col_id", middleware.IsSubscribed(), controller.CreateProject())
	project.Put("/:projectid", middleware.IsSubscribed(), controller.UpdateProject())
	project.Get("/", controller.GetAllProject())
	project.Get("/collectionProject/:col_id", controller.GetAllProjectOfCollectionId())
	project.Get("/readProject/:projectid",controller.GetAllProjectOfCollectionId())
	project.Delete("/deleteProject/:projectid",middleware.IsSubscribed(),controller.DeleteProject())
	project.Delete("/deleteAllProject/:col_id", middleware.IsSubscribed(), controller.DeleteAllProject())

	// Blog-section
	blog := app.Group("/api/v1/blog", middleware.ValidateAPIKey())

	blog.Post("/createBlog/:col_id",middleware.IsSubscribed(),controller.CreateBlog())
	blog.Get("/readAllBlog", controller.ReadBlog())
	blog.Get("/readAllBlogWithCol_id/:col_id", controller.ReadBlogWithCollectionId())
	blog.Get("/readOneBlog/:blogid", controller.ReadBlogWIthId())
	blog.Put("/updateBlog/:blogid", middleware.IsSubscribed(),controller.UpdateBlogWithBlogId())
	blog.Delete("/deleteBlog/:blogid", middleware.IsSubscribed(),controller.DeleteBlog())
	blog.Delete("/deleteAllBlog/:col_id", middleware.IsSubscribed(), controller.DeleteAllBlog())
	
	
	// Link section
	// this section can be made as your second brain
	link := app.Group("/api/v1/link", middleware.ValidateAPIKey())
	link.Post("/createLink", middleware.IsSubscribed(), controller.CreateLink())
	link.Get("/readLink",  controller.ReadLink())
	link.Get("/readLink/:linkid",  controller.ReadLinkWithLinkId())
	link.Put("/updateLink/:linkid", middleware.IsSubscribed(),controller.UpdateLinkWithLinkId())
	link.Delete("/deleteLink/:linkid", middleware.IsSubscribed(),controller.DeleteLinkWithLinkId())
	link.Delete("/deleteAllLink", middleware.IsSubscribed(), controller.DeleteAllLinks())

	media := app.Group("/api/media", middleware.ValidateAPIKey())
	media.Post("/postmedia/:col_id", controller.PostMedia())
	media.Get("/getALlMediaFiles/:col_id", controller.GetAllMediaFiles())
}

