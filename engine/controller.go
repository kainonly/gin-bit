package engine

import (
	"github.com/gin-gonic/gin"
	"net/http"
	"strconv"
)

type Controller struct {
	Engine  *Engine
	Service *Service
}

type CommonParams struct {
	Model string `uri:"model" binding:"omitempty,key"`
	Id    string `uri:"id" binding:"omitempty,objectId"`
}

func (x *Controller) Params(c *gin.Context) (params *CommonParams, err error) {
	if err = c.ShouldBindUri(&params); err != nil {
		return
	}
	if value, exists := c.Get(ModelNameKey); exists {
		params.Model = value.(string)
	}
	return
}

type CreateBody struct {
	Doc    map[string]interface{}   `json:"doc" binding:"required_without=Docs"`
	Docs   []map[string]interface{} `json:"docs" binding:"required_without=Doc,excluded_with=Doc,dive,gt=0"`
	Format map[string]interface{}   `json:"format" binding:"omitempty,dive,gt=0"`
	Ref    []string                 `json:"ref" binding:"omitempty,dive,gt=0"`
}

// Create 创建文档
func (x *Controller) Create(c *gin.Context) interface{} {
	params, err := x.Params(c)
	if err != nil {
		return err
	}
	var body CreateBody
	if err = c.ShouldBindJSON(&body); err != nil {
		return err
	}
	var result interface{}
	if len(body.Docs) != 0 {
		if result, err = x.Service.InsertMany(c.Request.Context(),
			params.Model, body.Docs, body.Format, body.Ref,
		); err != nil {
			return err
		}
	} else {
		if result, err = x.Service.InsertOne(c.Request.Context(),
			params.Model, body.Doc, body.Format, body.Ref,
		); err != nil {
			return err
		}
	}
	c.Set("status_code", http.StatusCreated)
	if err = x.Engine.Publish(params.Model, "create", EventValue{
		Body:     body,
		Response: result,
	}); err != nil {
		return err
	}
	return result
}

type FindQuery struct {
	Id     []string               `form:"id" binding:"omitempty,excluded_with=Where Single,dive,objectId"`
	Where  map[string]interface{} `form:"where" binding:"omitempty,excluded_with=Id"`
	Single bool                   `form:"single"`
	Sort   []string               `form:"sort" binding:"omitempty,dive,gt=0,sort"`
}

// Find 通过获取多个文档
func (x *Controller) Find(c *gin.Context) interface{} {
	params, err := x.Params(c)
	if err != nil {
		return err
	}
	var page Pagination
	if err = c.ShouldBindHeader(&page); err != nil {
		return err
	}
	var query FindQuery
	if err = c.ShouldBindQuery(&query); err != nil {
		return err
	}
	ctx := c.Request.Context()
	if query.Single == true {
		result, err := x.Service.FindOne(ctx, params.Model, query.Where)
		if err != nil {
			return err
		}
		return result
	}
	if len(query.Id) != 0 {
		result, err := x.Service.FindById(ctx, params.Model, query.Id, query.Sort)
		if err != nil {
			return err
		}
		return result
	}
	if page.Index != 0 && page.Size != 0 {
		result, err := x.Service.FindByPage(ctx, params.Model, page, query.Where, query.Sort)
		if err != nil {
			return err
		}
		c.Header("x-page-total", strconv.FormatInt(result.Total, 10))
		return result.Data
	}
	result, err := x.Service.Find(ctx, params.Model, query.Where, query.Sort)
	if err != nil {
		return err
	}
	return result
}

// FindOneById 通过 ID 获取单个文档
func (x *Controller) FindOneById(c *gin.Context) interface{} {
	params, err := x.Params(c)
	if err != nil {
		return err
	}
	result, err := x.Service.FindOneById(c.Request.Context(), params.Model, params.Id)
	if err != nil {
		return err
	}
	return result
}

type UpdateQuery struct {
	Id       []string               `form:"id" binding:"required_without=Where,excluded_with=Multiple,dive,objectId"`
	Where    map[string]interface{} `form:"where" binding:"required_without=Id,excluded_with=Id"`
	Multiple bool                   `form:"multiple"`
}

type UpdateBody struct {
	Update map[string]interface{} `json:"update" binding:"required"`
	Format map[string]interface{} `json:"format" binding:"omitempty,dive,gt=0"`
	Ref    []string               `json:"ref" binding:"omitempty,dive,gt=0"`
}

// Update 更新文档
func (x *Controller) Update(c *gin.Context) interface{} {
	params, err := x.Params(c)
	if err != nil {
		return err
	}
	var query UpdateQuery
	if err = c.ShouldBindQuery(&query); err != nil {
		return err
	}
	var body UpdateBody
	if err = c.ShouldBindJSON(&body); err != nil {
		return err
	}
	ctx := c.Request.Context()
	if len(query.Id) != 0 {
		result, err := x.Service.
			UpdateManyById(ctx, params.Model, query.Id, body.Update, body.Format, body.Ref)
		if err != nil {
			return err
		}
		return result
	}
	if query.Multiple {
		result, err := x.Service.
			UpdateMany(ctx, params.Model, query.Where, body.Update, body.Format, body.Ref)
		if err != nil {
			return err
		}
		return result
	}
	result, err := x.Service.
		UpdateOne(ctx, params.Model, query.Where, body.Update, body.Format, body.Ref)
	if err != nil {
		return err
	}
	if err = x.Engine.Publish(params.Model, "update", EventValue{
		Query:    query,
		Body:     body,
		Response: result,
	}); err != nil {
		return err
	}
	return result
}

func (x *Controller) UpdateOneById(c *gin.Context) interface{} {
	params, err := x.Params(c)
	if err != nil {
		return err
	}
	var body UpdateBody
	if err = c.ShouldBindJSON(&body); err != nil {
		return err
	}
	ctx := c.Request.Context()
	result, err := x.Service.
		UpdateOneById(ctx, params.Model, params.Id, body.Update, body.Format, body.Ref)
	if err != nil {
		return err
	}
	if err = x.Engine.Publish(params.Model, "update", EventValue{
		Id:       params.Id,
		Body:     body,
		Response: result,
	}); err != nil {
		return err
	}
	return result
}

type ReplaceOneBody struct {
	Doc    map[string]interface{} `json:"doc" binding:"required"`
	Format map[string]interface{} `json:"format" binding:"omitempty,dive,gt=0"`
	Ref    []string               `json:"ref" binding:"omitempty,dive,gt=0"`
}

func (x *Controller) ReplaceOneById(c *gin.Context) interface{} {
	params, err := x.Params(c)
	if err != nil {
		return err
	}
	var body ReplaceOneBody
	if err = c.ShouldBindJSON(&body); err != nil {
		return err
	}
	ctx := c.Request.Context()
	result, err := x.Service.ReplaceOneById(ctx, params.Model, params.Id, body.Doc, body.Format, body.Ref)
	if err != nil {
		return err
	}
	if err = x.Engine.Publish(params.Model, "replace", EventValue{
		Id:       params.Id,
		Body:     body,
		Response: result,
	}); err != nil {
		return err
	}
	return result
}

func (x *Controller) DeleteOneById(c *gin.Context) interface{} {
	params, err := x.Params(c)
	if err != nil {
		return err
	}
	ctx := c.Request.Context()
	result, err := x.Service.DeleteOneById(ctx, params.Model, params.Id)
	if err != nil {
		return err
	}
	if err = x.Engine.Publish(params.Model, "delete", EventValue{
		Id:       params.Id,
		Response: result,
	}); err != nil {
		return err
	}
	return result
}