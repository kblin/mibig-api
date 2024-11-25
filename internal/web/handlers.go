package web

import (
	"fmt"
	"net/http"

	"github.com/gin-gonic/gin"
	"github.com/spf13/viper"

	"secondarymetabolites.org/mibig-api/internal/data"
	"secondarymetabolites.org/mibig-api/internal/queries"
)

type VersionInfo struct {
	Api        string `json:"api"`
	BuildTime  string `json:"build_time"`
	GitVersion string `json:"git_version"`
}

func (app *application) version(c *gin.Context) {
	version_info := VersionInfo{
		Api:        "4.0alpha1",
		BuildTime:  viper.GetString("buildTime"),
		GitVersion: viper.GetString("gitVer"),
	}
	c.JSON(http.StatusOK, &version_info)
}

type Stats struct {
	Counts   *data.StatCounts   `json:"counts"`
	Clusters []data.StatCluster `json:"clusters"`
	Phyla    []data.TaxonStats  `json:"phyla"`
}

func (app *application) stats(c *gin.Context) {
	counts, err := app.Models.Entries.Counts()
	if err != nil {
		app.serverError(c, err)
		return
	}

	clusters, err := app.Models.Entries.ClusterStats()
	if err != nil {
		app.serverError(c, err)
		return
	}

	phyla, err := app.Models.Entries.PhylumStats()
	if err != nil {
		app.serverError(c, err)
		return
	}

	stat_info := Stats{
		Counts:   counts,
		Clusters: clusters,
		Phyla:    phyla,
	}

	c.JSON(http.StatusOK, &stat_info)
}

func (app *application) repository(c *gin.Context) {
	repository_entries, err := app.Models.Entries.Repository()
	if err != nil {
		app.serverError(c, err)
		return
	}

	c.JSON(http.StatusOK, repository_entries)
}

type queryContainer struct {
	Query        *queries.Query `json:"query"`
	SearchString string         `json:"search_string"`
	Paginate     int            `json:"paginate"`
	Offset       int            `json:"offset"`
	Verbose      bool           `json:"verbose"`
}

type queryResult struct {
	Total    int                    `json:"total"`
	Clusters []data.RepositoryEntry `json:"clusters"`
	Offset   int                    `json:"offset"`
	Paginate int                    `json:"paginate"`
	Stats    *data.ResultStats      `json:"stats"`
}

type queryError struct {
	Message string `json:"message"`
	Error   bool   `json:"error"`
}

func (app *application) search(c *gin.Context) {
	var qc queryContainer
	err := c.BindJSON(&qc)
	if err != nil {
		app.serverError(c, err)
		return
	}

	if qc.Query == nil && qc.SearchString == "" {
		c.JSON(http.StatusBadRequest, queryError{Message: "Invalid query", Error: true})
		return
	}

	if qc.Query == nil {
		qc.Query, err = queries.NewQueryFromString(qc.SearchString)
		if err != nil {
			app.serverError(c, err)
			return
		}
	}

	var entry_ids []string
	entry_ids, err = app.Models.Entries.Search(qc.Query.Terms)
	if err != nil {
		c.JSON(http.StatusBadRequest, queryError{Message: err.Error(), Error: true})
		return
	}

	var clusters []data.RepositoryEntry
	clusters, err = app.Models.Entries.Get(entry_ids)
	if err != nil {
		app.serverError(c, err)
		return
	}

	stats, err := app.Models.Entries.ResultStats(entry_ids)
	if err != nil {
		app.serverError(c, err)
		return
	}

	result := queryResult{
		Total:    len(entry_ids),
		Clusters: clusters,
		Offset:   qc.Offset,
		Paginate: qc.Paginate,
		Stats:    stats,
	}

	c.JSON(http.StatusOK, &result)
}

func (app *application) available(c *gin.Context) {
	category := c.Param("category")
	term := c.Param("term")
	available, err := app.Models.Entries.Available(category, term)
	if err == data.ErrInvalidCategory {
		c.JSON(http.StatusBadRequest, queryError{Message: err.Error(), Error: true})
		return
	} else if err != nil {
		app.serverError(c, err)
		return
	}

	c.JSON(http.StatusOK, &available)
}

func (app *application) Convert(c *gin.Context) {
	var req struct {
		Search string `form:"search_string"`
	}
	if err := c.Bind(&req); err != nil {
		c.JSON(http.StatusBadRequest, queryError{Message: err.Error(), Error: true})
		return
	}

	query, err := queries.NewQueryFromString(req.Search)
	if err != nil {
		app.serverError(c, err)
		return
	}

	err = app.Models.Entries.GuessCategories(query)
	if err == data.ErrInvalidCategory {
		c.JSON(http.StatusBadRequest, queryError{Message: err.Error(), Error: true})
		return
	} else if err != nil {
		app.serverError(c, err)
		return
	}

	c.JSON(http.StatusOK, query)
}

func (app *application) Contributors(c *gin.Context) {
	var req struct {
		Ids []string `form:"ids[]"`
	}
	if err := c.Bind(&req); err != nil {
		c.JSON(http.StatusBadRequest, queryError{Message: err.Error(), Error: true})
		return
	}

	contributors, err := app.Models.Entries.LookupContributors(req.Ids)
	if err != nil {
		app.serverError(c, err)
		return
	}

	c.JSON(http.StatusOK, contributors)
}

func (app *application) Redirect(c *gin.Context) {
	accession := c.Param("accession")
	entry, err := app.Models.Entries.Latest(accession)

	if err == data.ErrRecordNotFound {
		c.JSON(http.StatusNotFound, queryError{Message: err.Error(), Error: true})
	}
	if err != nil {
		app.serverError(c, err)
		return
	}
	target := fmt.Sprintf("/go/%s", entry.Accession)

	c.Redirect(http.StatusMovedPermanently, target)

}
