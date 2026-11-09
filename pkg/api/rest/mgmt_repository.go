package rest

import (
	"errors"
	"net/http"

	"github.com/cloud-barista/ai-ha/pkg/api/rest/model"
	"github.com/cloud-barista/ai-ha/pkg/framework/repository"
	"github.com/labstack/echo/v4"
	"github.com/rs/zerolog/log"
)

// RestGetRepositories godoc
// @ID GetRepositories
// @Summary List backup repositories
// @Description List the registered repositories, by name. Credentials are never returned.
// @Tags [Protection] Repository Management
// @Produce json
// @Success 200 {object} model.RepositoryListResp
// @Failure 401 {object} model.SimpleMsg
// @Failure 500 {object} model.SimpleMsg
// @Security AdminToken
// @Router /repositories [get]
func (h *mgmtHandler) RestGetRepositories(c echo.Context) error {
	list, err := h.repos.List(c.Request().Context())
	if err != nil {
		return refuse(c, err, http.StatusInternalServerError, model.CodeInternal, "Repositories could not be listed")
	}
	return c.JSON(http.StatusOK, model.RepositoryListResp{Items: list})
}

// RestPostRepository godoc
// @ID PostRepository
// @Summary Register a backup repository
// @Description Register an S3 repository (AWS or MinIO) that backups can write to. The credentials are stored encrypted, and an agent receives them only with a job it has claimed. The location cannot be changed afterwards: register a new repository for a new location. Whether the repository can be reached is not checked here.
// @Tags [Protection] Repository Management
// @Accept json
// @Produce json
// @Param repositoryReq body model.RepositoryReq true "Repository to register"
// @Success 201 {object} model.RepositoryInfo
// @Header 201 {string} Location "The repository's address"
// @Failure 400 {object} model.SimpleMsg
// @Failure 401 {object} model.SimpleMsg
// @Failure 409 {object} model.SimpleMsg
// @Failure 500 {object} model.SimpleMsg
// @Security AdminToken
// @Router /repositories [post]
func (h *mgmtHandler) RestPostRepository(c echo.Context) error {
	var req model.RepositoryReq
	if err := c.Bind(&req); err != nil {
		return refuseMalformed(c)
	}
	if err := c.Validate(&req); err != nil {
		return refuse(c, nil, http.StatusBadRequest, model.CodeInvalidRequest,
			"Name, provider (AWS or Minio), bucket, access key and secret key required")
	}

	created, err := h.repos.Create(c.Request().Context(), repository.Definition{
		Name:     req.Name,
		Provider: req.Provider,
		Endpoint: req.Endpoint,
		Region:   req.Region,
		Bucket:   req.Bucket,
		Prefix:   req.Prefix,
	}, repositoryCredentials(req.RepositoryCredentialsReq))
	switch {
	case errors.Is(err, repository.ErrInvalid):
		return refuse(c, err, http.StatusBadRequest, model.CodeInvalidRequest,
			"Repository rejected: check the name, endpoint, region, bucket and prefix")
	case errors.Is(err, repository.ErrExists):
		return refuse(c, nil, http.StatusConflict, model.CodeRepositoryExists, "A repository with this name already exists")
	case err != nil:
		return refuse(c, err, http.StatusInternalServerError, model.CodeInternal, "Repository could not be registered")
	}

	log.Info().Str("repositoryId", created.ID).Str("name", created.Name).
		Str("requestedBy", principalOf(c).Subject).Msg("Repository registered")
	c.Response().Header().Set(echo.HeaderLocation, "/aiha/repositories/"+created.ID)
	return c.JSON(http.StatusCreated, created)
}

// RestGetRepository godoc
// @ID GetRepository
// @Summary Get a backup repository
// @Description Return a registered repository, without its credentials.
// @Tags [Protection] Repository Management
// @Produce json
// @Param repoId path string true "Repository ID or name"
// @Success 200 {object} model.RepositoryInfo
// @Failure 401 {object} model.SimpleMsg
// @Failure 404 {object} model.SimpleMsg
// @Failure 500 {object} model.SimpleMsg
// @Security AdminToken
// @Router /repositories/{repoId} [get]
func (h *mgmtHandler) RestGetRepository(c echo.Context) error {
	info, err := h.repos.Get(c.Request().Context(), c.Param("repoId"))
	if errors.Is(err, repository.ErrNotFound) {
		return refuse(c, nil, http.StatusNotFound, model.CodeNotFound, "Repository not found")
	}
	if err != nil {
		return refuse(c, err, http.StatusInternalServerError, model.CodeInternal, "Repository could not be read")
	}
	return c.JSON(http.StatusOK, info)
}

// RestPutRepositoryCredentials godoc
// @ID PutRepositoryCredentials
// @Summary Replace a repository's credentials
// @Description Replace a repository's credentials, keeping its ID and location, so jobs and backup sets that name it stay valid. The old credentials are not read, which also recovers a repository whose secret key was lost.
// @Tags [Protection] Repository Management
// @Accept json
// @Produce json
// @Param repoId path string true "Repository ID or name"
// @Param repositoryCredentialsReq body model.RepositoryCredentialsReq true "New credentials"
// @Success 200 {object} model.RepositoryInfo
// @Failure 400 {object} model.SimpleMsg
// @Failure 401 {object} model.SimpleMsg
// @Failure 404 {object} model.SimpleMsg
// @Failure 500 {object} model.SimpleMsg
// @Security AdminToken
// @Router /repositories/{repoId}/credentials [put]
func (h *mgmtHandler) RestPutRepositoryCredentials(c echo.Context) error {
	var req model.RepositoryCredentialsReq
	if err := c.Bind(&req); err != nil {
		return refuseMalformed(c)
	}
	if err := c.Validate(&req); err != nil {
		return refuse(c, nil, http.StatusBadRequest, model.CodeInvalidRequest, "Access key and secret key required")
	}

	updated, err := h.repos.UpdateCredentials(c.Request().Context(), c.Param("repoId"), repositoryCredentials(req))
	switch {
	case errors.Is(err, repository.ErrNotFound):
		return refuse(c, nil, http.StatusNotFound, model.CodeNotFound, "Repository not found")
	case errors.Is(err, repository.ErrInvalid):
		return refuse(c, err, http.StatusBadRequest, model.CodeInvalidRequest, "Access key and secret key required")
	case err != nil:
		return refuse(c, err, http.StatusInternalServerError, model.CodeInternal, "Credentials could not be replaced")
	}

	log.Info().Str("repositoryId", updated.ID).Str("requestedBy", principalOf(c).Subject).
		Msg("Repository credentials replaced")
	return c.JSON(http.StatusOK, updated)
}

// RestPostRepositoryVerify godoc
// @ID PostRepositoryVerify
// @Summary Check that a repository can be reached
// @Description Not implemented yet: answers 501. A check from the framework says nothing about whether an agent inside a customer network can write there, so a real answer needs a job an agent carries out.
// @Tags [Protection] Repository Management
// @Accept json
// @Produce json
// @Param repoId path string true "Repository ID or name"
// @Param repositoryVerifyReq body model.RepositoryVerifyReq true "Who checks"
// @Success 202 {object} model.SimpleMsg
// @Failure 401 {object} model.SimpleMsg
// @Failure 501 {object} model.SimpleMsg
// @Security AdminToken
// @Router /repositories/{repoId}/verify [post]
func RestPostRepositoryVerify(c echo.Context) error {
	return notImplemented(c)
}

func repositoryCredentials(r model.RepositoryCredentialsReq) repository.Credentials {
	return repository.Credentials{AccessKey: r.AccessKey, SecretKey: r.SecretKey, SessionToken: r.SessionToken}
}
