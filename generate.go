package main

import (
	"fmt"
	"log"
	"os"
	"strings"

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

func main() {
	// Retrieve values from environment variables
	orgName := os.Getenv("PLUGIN_ORG_NAME")
	projectName := os.Getenv("PLUGIN_PROJECT_NAME")
	projectColor := os.Getenv("PLUGIN_PROJECT_COLOR") // TODO: Use random hex generator
	orgID := strings.ToLower(orgName)                 // Assuming organization ID is set as an environment variable

	// Create Organization Module
	organization := Organization{
		Name:   orgName,
		Tags:   map[string]string{"bu": orgName},
		Source: "harness-community/structure/harness//modules/organizations",
	}

	file := hclwrite.NewEmptyFile()
	rootBody := file.Body()

	// Set up the organization block
	organizationBlock := rootBody.AppendNewBlock("module", []string{"organization"})
	orgBody := organizationBlock.Body()
	orgBody.SetAttributeValue("name", cty.StringVal(organization.Name))
	orgBody.SetAttributeValue("source", cty.StringVal(organization.Source))
	// Add tags as needed

	// Create the project module
	project := Project{
		Name:           projectName,
		OrganizationID: orgID,
		Color:          projectColor,
		Tags:           map[string]string{"bu": orgID, "app": "ApplicationA"},
		Source:         "harness-community/structure/harness//modules/projects",
	}

	// Set up the project block
	projectBlock := rootBody.AppendNewBlock("module", []string{"project"})
	projectBody := projectBlock.Body()
	projectBody.SetAttributeValue("name", cty.StringVal(project.Name))
	projectBody.SetAttributeValue("organization_id", cty.StringVal(project.OrganizationID))
	projectBody.SetAttributeValue("color", cty.StringVal(project.Color))
	projectBody.SetAttributeValue("source", cty.StringVal(project.Source))
	// Add tags as needed

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
