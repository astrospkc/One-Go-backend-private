package controller

import (
	"context"
	"encoding/json"
	"fmt"
	"strings"

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

			Build target (STRICT):
			- The project MUST be statically buildable
			- The output MUST work with static hosting
			- The final build output MUST be HTML, CSS, and client-side JavaScript only
			- The site MUST NOT require a backend server at runtime

			Tech stack (STRICT):
			- React 19
			- TypeScript
			- Tailwind CSS
			- Vite (static build) OR Next.js with static export ONLY

			Static constraints (VERY IMPORTANT):
			- Do NOT use server components
			- Do NOT use API routes
			- Do NOT use middleware
			- Do NOT use environment variables
			- Do NOT use filesystem, process, or child_process
			- Do NOT use dynamic imports
			- Do NOT use SSR or runtime rendering
			- Do NOT assume any backend

			SEO rules:
			- Use semantic HTML
			- Add meta tags
			- Ensure all pages are statically rendered

			Design requirements:
			- Clean, modern design
			- Mobile responsive
			- Accessible layout

			Project data (JSON):
			%s

			User request:
			%s

			Output rules (STRICT):
			- Output CODE ONLY
			- Do NOT include explanations or markdown
			- Use comments to separate files
			- Use EXACT format:
			// FILE: path/to/file

			Failure handling:
			- If any feature cannot be implemented within these constraints,
			OMIT the feature instead of breaking the build.

			IMPORTANT:
			Before outputting, internally verify that the project builds successfully.
			If it would fail, fix the code before outputting.

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
		
		result:=resp.Text()
		files:=ParseFiles(result)
		
		fmt.Println(resp.ExecutableCode())
		fmt.Println(resp.CodeExecutionResult(), files)
		


		return c.Status(fiber.StatusOK).JSON(fiber.Map{
			"success": true,
			"data":    nil,
		})
	}
}

type GeneratedFle struct{
	Path string
	Content string
}

// siteId := uuid.NewString()

func ParseFiles(aiOutput string)[]GeneratedFle{
	var files []GeneratedFle

	parts := strings.Split(aiOutput, "")
	for _,part:=range parts{
		part = strings.TrimSpace(part)
		if part==""{
			continue
		}
		lines:=strings.SplitN(part,"\n" ,2)
		if len(lines)<2{
			continue
		}

		files = append(files, GeneratedFle{
			Path: strings.TrimSpace(lines[0]),
			Content:strings.TrimSpace(lines[1]),
		})
	}
	return files
}

// generate file presigned url
