package routes

import (
	"gobackend/controller"
	"gobackend/middleware"

	"github.com/gofiber/fiber/v2"
)



func RegisterNormalRoutes(app *fiber.App){
		// handling normal routings

	auth:= app.Group("/auth")
	auth.Post("/register/send-otp", controller.RegisterSendOtp())
	auth.Post("/register/verify-otp", controller.RegisterVerifyOTP())
	auth.Post("/login", controller.Login())
	auth.Post("/logout", controller.Logout())
	auth.Get("/user",middleware.FetchUser(), controller.GetUser())
	auth.Put("/editUser", middleware.FetchUser(), controller.UpdateUser())
	auth.Post("/forgot-password", controller.ForgotPassword())
	auth.Post("/reset-password", controller.ResetPassword())

	// handling collections routes
	col := app.Group("/collection",middleware.FetchUser())
	col.Post("/", controller.CreateCollection())
	col.Get("/", controller.GetAllCollection())
	col.Delete("/deleteAllCollection", controller.DeleteAllCollection())
	col.Delete("/:id", controller.DeleteCollection())
	col.Put("/:id", controller.UpdateCollection())
	col.Get("/:id", controller.GetCollection())
	
	// handling normal routings with auth middleware
	project := app.Group("/project", middleware.FetchUser())
	project.Post("/presignedUrl", controller.GetFilePresignedUrl())
	
	project.Post("createProject/:col_id", controller.CreateProject())
	project.Put("/:projectid", controller.UpdateProject())
	project.Get("/", controller.GetAllProject())
	project.Get("/collectionProject/:col_id", controller.GetAllProjectOfCollectionId())
	project.Get("/readProject/:projectid",controller.FindOneViaPID())
	project.Delete("/deleteProject/:projectid",controller.DeleteProject())
	project.Delete("/deleteAllProject/:col_id", controller.DeleteAllProject())

	// Blog-section
	blog := app.Group("/blog", middleware.FetchUser())

	blog.Post("/createBlog/:col_id",controller.CreateBlog())
	blog.Get("/readAllBlog", controller.ReadBlog())
	blog.Get("/readAllBlogWithCol_id/:col_id", controller.ReadBlogWithCollectionId())
	blog.Get("/readOneBlog/:blogid", controller.ReadBlogWIthId())
	blog.Put("/updateBlog/:blogid",controller.UpdateBlogWithBlogId())
	blog.Delete("/deleteBlog/:blogid", controller.DeleteBlog())
	blog.Delete("/deleteAllBlog/:col_id", controller.DeleteAllBlog())
	
	
	// Link section
	// this section can be made as your second brain
	link := app.Group("/link", middleware.FetchUser())
	link.Post("/createLink",  controller.CreateLink())
	link.Get("/readLink",  controller.ReadLink())
	link.Get("/readLink/:linkid",  controller.ReadLinkWithLinkId())
	link.Put("/updateLink/:linkid", controller.UpdateLinkWithLinkId())
	link.Delete("/deleteLink/:linkid", controller.DeleteLinkWithLinkId())
	link.Delete("/deleteAllLink", controller.DeleteAllLinks())

	media := app.Group("/media", middleware.FetchUser())
	media.Post("/postmedia/:col_id", controller.PostMedia())
	media.Get("/getALlMediaFiles/:col_id", controller.GetAllMediaFiles())
	media.Post("/showMediaFiles/:col_id",controller.ShowFile())
	media.Delete("/deleteMedia/:media_id", controller.DeleteFile())
}