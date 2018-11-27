package handlers

import (
	"context"
	"fmt"
	"net/http"
	"strings"

	"go-scim/shared"
)

func CreateGroupHandler(r shared.WebRequest, server ScimServer, ctx context.Context) (ri *ResponseInfo) {
	ri = newResponse()
	sch := server.InternalSchema(shared.GroupUrn)

	resource, err := ParseBodyAsResource(r)
	ErrorCheck(err)

	err = server.ValidateType(resource, sch, ctx)
	ErrorCheck(err)

	err = server.CorrectCase(resource, sch, ctx)
	ErrorCheck(err)

	err = server.ValidateRequired(resource, sch, ctx)
	ErrorCheck(err)

	repo := server.Repository(shared.GroupResourceType)
	err = server.ValidateUniqueness(resource, sch, repo, ctx)
	ErrorCheck(err)

	err = server.AssignReadOnlyValue(resource, ctx)
	ErrorCheck(err)

	err = repo.Create(resource)
	ErrorCheck(err)

	json, err := server.MarshalJSON(resource, sch, []string{}, []string{})
	ErrorCheck(err)

	location := resource.GetData()["meta"].(map[string]interface{})["location"].(string)
	version := resource.GetData()["meta"].(map[string]interface{})["version"].(string)

	ri.Status(http.StatusCreated)
	ri.ScimJsonHeader()
	if len(version) > 0 {
		ri.ETagHeader(version)
	}
	if len(location) > 0 {
		ri.LocationHeader(location)
	}
	ri.Body(json)
	return
}

func PatchGroupHandler(r shared.WebRequest, server ScimServer, ctx context.Context) (ri *ResponseInfo) {
	ri = newResponse()
	sch := server.InternalSchema(shared.GroupUrn)
	repo := server.Repository(shared.GroupResourceType)

	id, version := ParseIdAndVersion(r)
	ctx = context.WithValue(ctx, shared.ResourceId{}, id)

	resource, err := repo.Get(id, version)
	ErrorCheck(err)

	mod, err := ParseModification(r)
	ErrorCheck(err)
	err = mod.Validate()
	ErrorCheck(err)

	for _, patch := range mod.Ops {
		err = server.ApplyPatch(patch, resource.(*shared.Resource), sch, ctx)
		ErrorCheck(err)
	}

	reference, err := repo.Get(id, version)
	ErrorCheck(err)

	err = server.ValidateType(resource.(*shared.Resource), sch, ctx)
	ErrorCheck(err)

	err = server.CorrectCase(resource.(*shared.Resource), sch, ctx)
	ErrorCheck(err)

	err = server.ValidateRequired(resource.(*shared.Resource), sch, ctx)
	ErrorCheck(err)

	err = server.ValidateMutability(resource.(*shared.Resource), reference.(*shared.Resource), sch, ctx)
	ErrorCheck(err)

	err = server.ValidateUniqueness(resource.(*shared.Resource), sch, repo, ctx)
	ErrorCheck(err)

	err = server.AssignReadOnlyValue(resource.(*shared.Resource), ctx)
	ErrorCheck(err)

	err = repo.Update(id, version, resource)
	ErrorCheck(err)

	json, err := server.MarshalJSON(resource, sch, []string{}, []string{})
	ErrorCheck(err)

	location := resource.GetData()["meta"].(map[string]interface{})["location"].(string)
	newVersion := resource.GetData()["meta"].(map[string]interface{})["version"].(string)

	ri.Status(http.StatusOK)
	ri.ScimJsonHeader()
	if len(newVersion) > 0 {
		ri.ETagHeader(newVersion)
	}
	if len(location) > 0 {
		ri.LocationHeader(location)
	}
	ri.Body(json)
	return
}

func ReplaceGroupHandler(r shared.WebRequest, server ScimServer, ctx context.Context) (ri *ResponseInfo) {
	ri = newResponse()
	sch := server.InternalSchema(shared.GroupUrn)
	repo := server.Repository(shared.GroupResourceType)

	resource, err := ParseBodyAsResource(r)
	ErrorCheck(err)

	//id, version := ParseIdAndVersion(r)

	// BEGIN NO ID HACK
	parts := strings.Split(r.Target(), "/")
	id := parts[len(parts)-1]
	if resource.Complex["id"] == nil {
		resource.Complex["id"] = id
	}
	// END NO ID HACK

	// BEGIN MEMBERS FIX
	// If members was not present in the request, it should be interpreted as an empty array
	// because it is required=false and this is a PUT request which replaces a resource in its
	// entirety.
	// Note that both the PUT and PATCH call upon repo.UpdateGroup() and the above only holds for
	// PUT requests, so we have to update the empty members array here and not in repo.UpdateGroup()
	if _, found := resource.Complex["members"]; !found {
		resource.Complex["members"] = make([]interface{},0)
	}
	// END MEMBERS FIX

	version := ""
	ctx = context.WithValue(ctx, shared.ResourceId{}, id)

	//reference, err := repo.Get(id, version)
	//ErrorCheck(err)

	err = server.ValidateType(resource, sch, ctx)
	ErrorCheck(err)

	err = server.CorrectCase(resource, sch, ctx)
	ErrorCheck(err)
	err = server.ValidateRequired(resource, sch, ctx)
	ErrorCheck(err)

	// err = server.ValidateMutability(resource, reference.(*shared.Resource), sch, ctx)
	// ErrorCheck(err)
	err = server.ValidateUniqueness(resource, sch, repo, ctx)
	ErrorCheck(err)

	err = server.AssignReadOnlyValue(resource, ctx)
	ErrorCheck(err)

	err = repo.Update(id, version, resource)
	ErrorCheck(err)

	json, err := server.MarshalJSON(resource, sch, []string{}, []string{})
	ErrorCheck(err)

	location := resource.GetData()["meta"].(map[string]interface{})["location"].(string)
	newVersion := resource.GetData()["meta"].(map[string]interface{})["version"].(string)

	ri.Status(http.StatusOK)
	ri.ScimJsonHeader()
	if len(newVersion) > 0 {
		ri.ETagHeader(newVersion)
	}
	if len(location) > 0 {
		ri.LocationHeader(location)
	}
	ri.Body(json)

	return
}

func QueryGroupHandler(r shared.WebRequest, server ScimServer, ctx context.Context) (ri *ResponseInfo) {
	ri = newResponse()
	sch := server.InternalSchema(shared.GroupUrn)

	attributes, excludedAttributes := ParseInclusionAndExclusionAttributes(r)

	sr, err := ParseSearchRequest(r, server)
	ErrorCheck(err)

	err = sr.Validate(sch)
	ErrorCheck(err)

	repo := server.Repository(shared.GroupResourceType)
	lr, err := repo.Search(sr)
	ErrorCheck(err)

	json, err := server.MarshalJSON(lr, sch, attributes, excludedAttributes)
	ErrorCheck(err)

	ri.Status(http.StatusOK)
	ri.ScimJsonHeader()
	ri.Body(json)
	return
}

func DeleteGroupByIdHandler(r shared.WebRequest, server ScimServer, ctx context.Context) (ri *ResponseInfo) {
	ri = newResponse()

	id, version := ParseIdAndVersion(r)
	repo := server.Repository(shared.GroupResourceType)

	err := repo.Delete(id, version)
	ErrorCheck(err)

	ri.Status(http.StatusNoContent)
	return
}

func GetGroupByIdHandler(r shared.WebRequest, server ScimServer, ctx context.Context) (ri *ResponseInfo) {
	ri = newResponse()
	sch := server.InternalSchema(shared.GroupUrn)

	id, version := ParseIdAndVersion(r)

	if len(version) > 0 {
		count, err := server.Repository(shared.GroupResourceType).Count(
			fmt.Sprintf("id eq \"%s\" and meta.version eq \"%s\"", id, version),
		)
		if err == nil && count > 0 {
			ri.Status(http.StatusNotModified)
			return
		}
	}

	attributes, excludedAttributes := ParseInclusionAndExclusionAttributes(r)

	dp, err := server.Repository(shared.GroupResourceType).Get(id, version)
	ErrorCheck(err)
	location := dp.GetData()["meta"].(map[string]interface{})["location"].(string)

	json, err := server.MarshalJSON(dp, sch, attributes, excludedAttributes)
	ErrorCheck(err)

	ri.Status(http.StatusOK)
	ri.ScimJsonHeader()
	if len(version) > 0 {
		ri.ETagHeader(version)
	}
	if len(location) > 0 {
		ri.LocationHeader(location)
	}
	ri.Body(json)
	return
}
