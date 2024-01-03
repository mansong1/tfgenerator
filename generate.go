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
