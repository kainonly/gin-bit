package api

import (
	"github.com/gin-gonic/gin"
	"github.com/weplanx/go/route"
	"go.mongodb.org/mongo-driver/bson"
	"net/http"
	"strconv"
)

type Controller struct {
	Service *Service
}

func (x *Controller) Auto(r *gin.Engine) {
	r.POST("/:name", route.Use(x.Create))
	r.GET("/:name", route.Use(x.Find))
	r.GET("/:name/:id", route.Use(x.FindOneById))
	r.PATCH("/:name", route.Use(x.Update))
	r.PATCH("/:name/:id", route.Use(x.UpdateOneById))
	r.PUT("/:name/:id", route.Use(x.ReplaceOneById))
	r.DELETE("/:name/:id", route.Use(x.DeleteOneById))
}

type CommonParams struct {
	Name string `uri:"name" binding:"required,key"`
	Id   string `uri:"id" binding:"omitempty,objectId"`
}

type CreateBody struct {
	Doc bson.M `json:"doc" binding:"required"`
}

// Create 创建文档
func (x *Controller) Create(c *gin.Context) interface{} {
	var params CommonParams
	if err := c.ShouldBindUri(&params); err != nil {
		return err
	}
	var body CreateBody
	if err := c.ShouldBindJSON(&body); err != nil {
		return err
	}
	result, err := x.Service.Create(c.Request.Context(), params.Name, body.Doc)
	if err != nil {
		return err
	}
	c.Set("status_code", http.StatusCreated)
	return result
}

type FindQuery struct {
	Id     []string `form:"id" binding:"omitempty,excluded_with=Where Single,dive,objectId"`
	Where  bson.M   `form:"where" binding:"omitempty,excluded_with=Id"`
	Single bool     `form:"single"`
	Sort   []string `form:"sort" binding:"omitempty,dive,gt=0,sort"`
}

// Find 通过获取多个文档
func (x *Controller) Find(c *gin.Context) interface{} {
	var params CommonParams
	if err := c.ShouldBindUri(&params); err != nil {
		return err
	}
	var page Pagination
	if err := c.ShouldBindHeader(&page); err != nil {
		return err
	}
	var query FindQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		return err
	}
	ctx := c.Request.Context()
	if query.Single == true {
		result, err := x.Service.FindOne(ctx, params.Name, query.Where)
		if err != nil {
			return err
		}
		return result
	}
	if len(query.Id) != 0 {
		result, err := x.Service.FindById(ctx, params.Name, query.Id, query.Sort)
		if err != nil {
			return err
		}
		return result
	}
	if page.Index != 0 && page.Size != 0 {
		result, err := x.Service.FindByPage(ctx, params.Name, page, query.Where, query.Sort)
		if err != nil {
			return err
		}
		c.Header("x-page-total", strconv.FormatInt(result.Total, 10))
		return result.Data
	}
	result, err := x.Service.Find(ctx, params.Name, query.Where, query.Sort)
	if err != nil {
		return err
	}
	return result
}

// FindOneById 通过 ID 获取单个文档
func (x *Controller) FindOneById(c *gin.Context) interface{} {
	var params CommonParams
	if err := c.ShouldBindUri(&params); err != nil {
		return err
	}
	err, result := x.Service.FindOneById(c.Request.Context(), params.Name, params.Id)
	if err != nil {
		return err
	}
	return result
}

type UpdateQuery struct {
	Id       []string `form:"id" binding:"required_without=Where,excluded_with=Multiple,dive,objectId"`
	Where    bson.M   `form:"where" binding:"required_without=Id,excluded_with=Id"`
	Multiple bool     `form:"multiple"`
}

type UpdateBody struct {
	Update bson.M `json:"update" binding:"required"`
}

// Update 更新文档
func (x *Controller) Update(c *gin.Context) interface{} {
	var params CommonParams
	if err := c.ShouldBindUri(&params); err != nil {
		return err
	}
	var query UpdateQuery
	if err := c.ShouldBindQuery(&query); err != nil {
		return err
	}
	var body UpdateBody
	if err := c.ShouldBindJSON(&body); err != nil {
		return err
	}
	ctx := c.Request.Context()
	if len(query.Id) != 0 {
		result, err := x.Service.
			UpdateManyById(ctx, params.Name, query.Id, body.Update)
		if err != nil {
			return err
		}
		return result
	}
	if query.Multiple {
		result, err := x.Service.
			UpdateMany(ctx, params.Name, query.Where, body.Update)
		if err != nil {
			return err
		}
		return result
	}
	result, err := x.Service.
		UpdateOne(ctx, params.Name, query.Where, body.Update)
	if err != nil {
		return err
	}
	return result
}

func (x *Controller) UpdateOneById(c *gin.Context) interface{} {
	var params CommonParams
	if err := c.ShouldBindUri(&params); err != nil {
		return err
	}
	var body UpdateBody
	if err := c.ShouldBindJSON(&body); err != nil {
		return err
	}
	ctx := c.Request.Context()
	result, err := x.Service.
		UpdateOneById(ctx, params.Name, params.Id, body.Update)
	if err != nil {
		return err
	}
	return result
}

type ReplaceOneBody struct {
	Doc bson.M `json:"doc" binding:"required"`
}

func (x *Controller) ReplaceOneById(c *gin.Context) interface{} {
	var params CommonParams
	if err := c.ShouldBindUri(&params); err != nil {
		return err
	}
	var body ReplaceOneBody
	if err := c.ShouldBindJSON(&body); err != nil {
		return err
	}
	ctx := c.Request.Context()
	result, err := x.Service.ReplaceOneById(ctx, params.Name, params.Id, body.Doc)
	if err != nil {
		return err
	}
	return result
}

func (x *Controller) DeleteOneById(c *gin.Context) interface{} {
	var params CommonParams
	if err := c.ShouldBindUri(&params); err != nil {
		return err
	}
	ctx := c.Request.Context()
	result, err := x.Service.DeleteOneById(ctx, params.Name, params.Id)
	if err != nil {
		return err
	}
	return result
}