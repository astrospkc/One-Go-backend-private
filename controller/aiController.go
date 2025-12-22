package controller

import (
	"context"
	"encoding/json"
	"fmt"

	"time"

	"gobackend/connect"
	"gobackend/models"

	"go.mongodb.org/mongo-driver/v2/bson"
	"google.golang.org/genai"

	"github.com/gofiber/fiber/v2"
)



type APIProject struct {
	Id           string    `bson:"id,omitempty" json:"id"`
	Title        string    `bson:"title" json:"title"`
	Description  string    `bson:"description,omitempty" json:"description"`
	Tags         string    `bson:"tags,omitempty" json:"tags"`
	FileUpload   []string  `bson:"fileUpload,omitempty" json:"fileUpload"`
	Thumbnail    string    `bson:"thumbnail,omitempty" json:"thumbnail"`
	GithubLink   string    `bson:"githublink,omitempty" json:"githublink"`
	DemoLink     string    `bson:"demolink,omitempty" json:"demolink"`
	LiveDemoLink string    `bson:"livedemolink,omitempty" json:"livedemolink"`
	BlogLink     string    `bson:"blogLink,omitempty" json:"blogLink"`
	TeamMembers  string    `bson:"teamMembers,omitempty" json:"teamMembers"`
	CreatedAt    time.Time `bson:"created_at" json:"created_at"`
	UpdatedAt    time.Time `bson:"updated_at" json:"updated_at"`
}

func GenerateAIContent() fiber.Handler {
	return func(c *fiber.Ctx) error {
		col_id:=c.Params("col_id")
		fmt.Println("col_id", col_id)
		var body struct {
			Content string `json:"content"`
		}

		if err := c.BodyParser(&body); err != nil {
			return c.Status(fiber.StatusBadRequest).JSON(fiber.Map{
				"success": false,
				"data":    nil,
			})
		}

		// get the collection id from the request
		// get all the projects of the collection 
		var projects []models.Project 
		cursor,err:= connect.ProjectCollection.Find(context.TODO(), bson.M{"collection_id": col_id})
		if err := cursor.All(context.TODO(), &projects); err != nil {
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"data":    nil,
			})
		} 
		
		apiProjects:= []APIProject{}
		for _, project := range projects {
			apiProjects = append(apiProjects, APIProject{
				Id:           project.Id,
				Title:        project.Title,
				Description:  project.Description,
				Tags:         project.Tags,
				FileUpload:   project.FileUpload,
				Thumbnail:    project.Thumbnail,
				GithubLink:   project.GithubLink,
				DemoLink:     project.DemoLink,
				LiveDemoLink: project.LiveDemoLink,
				BlogLink:     project.BlogLink,
				TeamMembers:  project.TeamMembers,
				CreatedAt:    project.CreatedAt,
				UpdatedAt:    project.UpdatedAt,
			})
		}

		fmt.Println("projects: ", projects)

		projectJson,_:=json.MarshalIndent(apiProjects,""," ")

		prompt := fmt.Sprintf(`
			You are a senior frontend engineer.

			Your task:
			Generate a complete developer portfolio website using the provided project data.

			Constraints:
			- Use Next.js 14 (App Router)
			- TypeScript
			- Tailwind CSS
			- Clean, modern design
			- SEO friendly
			- Mobile responsive

			Project data (JSON):
			%s

			User request:
			%s

			Output rules:
			- Output CODE ONLY
			- Use comments to separate files
			- Use format: // FILE: path/to/file

			Generate the full portfolio now.
			`, string(projectJson), body.Content)


		config := &genai.GenerateContentConfig{
        Tools: []*genai.Tool{
            {CodeExecution: &genai.ToolCodeExecution{}},
        },
    }
		model := "gemini-2.5-flash"
		resp, err:=connect.GeminiClient.Models.GenerateContent(context.TODO(),model,genai.Text(prompt), config)
		if err!=nil{
			return c.Status(fiber.StatusInternalServerError).JSON(fiber.Map{
				"success": false,
				"data":    nil,
			})
		}
		
		fmt.Println(resp.Text())
		fmt.Println(resp.ExecutableCode())
		fmt.Println(resp.CodeExecutionResult())
		


		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"data":    nil,
		})
	}
}