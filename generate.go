package main

import (
	"fmt"
	"log"
	"math/rand"
	"os"
	"strings"
	"time"

	"github.com/hashicorp/hcl/v2"
	"github.com/hashicorp/hcl/v2/hclwrite"
	"github.com/zclconf/go-cty/cty"
)

type Organization struct {
	Name   string
	Tags   map[string]string
	Source string
}

type Project struct {
	Name           string
	OrganizationID string
	Color          string
	Tags           map[string]string
	Source         string
}

// generateRandomHexColor generates a random hex color string
func generateRandomHexColor() string {
	source := rand.NewSource(time.Now().UnixNano())
	random := rand.New(source)
	return fmt.Sprintf("#%06x", random.Intn(16777215)) // 16777215 is FFFFFF in decimal
}

func main() {
	// Retrieve values from environment variables
	orgName := os.Getenv("PLUGIN_ORG_NAME")
	githubRepo := os.Getenv("PLUGIN_GITHUB_REPO")
	projectName := os.Getenv("PLUGIN_PROJECT_NAME")
	projectColorEnv := os.Getenv("PLUGIN_PROJECT_COLOR")

	var projectColor string
	if projectColorEnv != "" {
		projectColor = projectColorEnv
	} else {
		projectColor = generateRandomHexColor()
	}
	orgID := strings.ToLower(orgName) // Assuming organization ID is set as an environment variable

	// Create Organization Module
	organization := Organization{
		Name:   orgName,
		Tags:   map[string]string{"bu": orgName},
		Source: "harness-community/structure/harness//modules/organizations",
	}

	file := hclwrite.NewEmptyFile()
	rootBody := file.Body()

	// Set up the organization block
	organizationBlock := rootBody.AppendNewBlock("module", []string{"organization_" + orgID})
	orgBody := organizationBlock.Body()
	orgBody.SetAttributeValue("name", cty.StringVal(organization.Name))
	orgBody.SetAttributeValue("source", cty.StringVal(organization.Source))
	//Add tags as needed
	orgTagsMap := make(map[string]cty.Value)
	for key, value := range organization.Tags {
		orgTagsMap[key] = cty.StringVal(value)
	}
	orgTags := cty.MapVal(orgTagsMap)
	orgBody.SetAttributeValue("tags", orgTags)

	// Create the project module
	project := Project{
		Name:           projectName,
		OrganizationID: orgID,
		Color:          projectColor,
		Tags:           map[string]string{"bu": orgID, "app": "ApplicationA"},
		Source:         "harness-community/structure/harness//modules/projects",
	}

	// Set up the project block
	projectBlock := rootBody.AppendNewBlock("module", []string{"project_" + projectName})
	projectBody := projectBlock.Body()
	projectBody.SetAttributeValue("name", cty.StringVal(project.Name))
	projectBody.SetAttributeTraversal("organization_id", hcl.Traversal{
		hcl.TraverseRoot{
			Name: "module.organization_" + orgID,
		},
		hcl.TraverseAttr{
			Name: "organization_details",
		},
		hcl.TraverseAttr{
			Name: "id",
		},
	})
	projectBody.SetAttributeValue("color", cty.StringVal(project.Color))
	projectBody.SetAttributeValue("source", cty.StringVal(project.Source))
	// Add tags as needed
	projectTagsMap := make(map[string]cty.Value)
	for key, value := range project.Tags {
		projectTagsMap[key] = cty.StringVal(value)
	}
	projectTags := cty.MapVal(projectTagsMap)
	projectBody.SetAttributeValue("tags", projectTags)

	// Create the project module
	templatesBlock := rootBody.AppendNewBlock("module", []string{"hello_world_template_" + projectName})
	templatesBody := templatesBlock.Body()
	templatesBody.SetAttributeValue("source", cty.StringVal("harness-community/content/harness//modules/templates"))
	templatesBody.SetAttributeValue("name", cty.StringVal("Welcome to Harness"))
	templatesBody.SetAttributeTraversal("organization_id", hcl.Traversal{
		hcl.TraverseRoot{
			Name: "module.organization_" + orgID,
		},
		hcl.TraverseAttr{
			Name: "organization_details",
		},
		hcl.TraverseAttr{
			Name: "id",
		},
	})
	templatesBody.SetAttributeTraversal("project_id", hcl.Traversal{
		hcl.TraverseRoot{
			Name: "module.project_" + projectName,
		},
		hcl.TraverseAttr{
			Name: "project_details",
		},
		hcl.TraverseAttr{
			Name: "id",
		},
	})

	// Set up yaml_data with templatefile function
	yamlData := fmt.Sprintf(`templatefile("${path.module}/templates/templates/welcome-to-harness.yaml", { REPOSITORY_NAME : "%s" })`, githubRepo)
	templatesBody.SetAttributeValue("yaml_data", cty.StringVal(yamlData))

	templatesBody.SetAttributeValue("template_version", cty.StringVal("v1.0.0"))
	templatesBody.SetAttributeValue("type", cty.StringVal("Pipeline"))

	// Add tags
	templatesTagsMap := make(map[string]cty.Value)
	for key, value := range project.Tags {
		templatesTagsMap[key] = cty.StringVal(value)
	}
	templatesTags := cty.MapVal(templatesTagsMap)
	templatesBody.SetAttributeValue("tags", templatesTags)

	// Write the file
	tfFileName := "main_" + orgName + ".tf"
	tfFile, err := os.Create(tfFileName)
	if err != nil {
		log.Fatal(err)
	}
	defer tfFile.Close()

	_, err = tfFile.Write(file.Bytes())
	fmt.Printf("Writing out %s for Org %s and Project %s\n", tfFileName, orgName, projectName)
	if err != nil {
		log.Fatal(err)
	}
}
