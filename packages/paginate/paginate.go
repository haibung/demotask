package paginate

import (
	"fmt"
	"github.com/labstack/echo/v4"
	"gorm.io/gorm"
	"math"
	"strconv"
)

type Pagination struct {
	Limit      int    `json:"limit"`
	Page       int    `json:"page"`
	TotalRows  int64  `json:"total_rows"`
	TotalPages int    `json:"total_pages"`
	NextPage   string `json:"next_page"`
	PrevPage   string `json:"prev_page"`
}

// GetOffset :
func (receiver *Pagination) GetOffset() int {
	return (receiver.GetPage() - 1) * receiver.GetLimit()
}

// GetLimit : Default limit 10
func (receiver *Pagination) GetLimit() int {
	if receiver.Limit == 0 {
		receiver.Limit = 10
	}
	return receiver.Limit
}

// GetPage : Default page 1
func (receiver *Pagination) GetPage() int {
	if receiver.Page == 0 {
		receiver.Page = 1
	}
	return receiver.Page
}

// Next :
func (receiver *Pagination) Next(c echo.Context) string {
	if receiver.GetPage() <= receiver.TotalPages || receiver.GetPage() >= receiver.TotalPages {
		if receiver.GetPage() >= receiver.TotalPages && receiver.TotalPages == 0 {
			receiver.NextPage = receiver.PerPage(c, 1)
		} else if receiver.GetPage() >= receiver.TotalPages && receiver.TotalPages > 0 {
			receiver.NextPage = receiver.PerPage(c, receiver.TotalPages)
		} else {
			receiver.NextPage = receiver.PerPage(c, receiver.GetPage()+1)
		}
	} else {
		receiver.NextPage = receiver.PerPage(c, 1)
	}

	return receiver.NextPage
}

// Prev :
func (receiver *Pagination) Prev(c echo.Context) string {
	if receiver.GetPage() > 1 {
		if receiver.GetPage() > receiver.TotalPages {
			prevPage := receiver.TotalPages - 1
			if prevPage < 1 {
				prevPage = 1
			}
			receiver.PrevPage = receiver.PerPage(c, prevPage)
		} else {
			receiver.PrevPage = receiver.PerPage(c, receiver.GetPage()-1)
		}
	} else {
		receiver.PrevPage = receiver.PerPage(c, 1)
	}

	return receiver.PrevPage
}

// PerPage :
func (receiver *Pagination) PerPage(c echo.Context, page int) string {
	c.QueryParams().Set("page", strconv.Itoa(page))
	url := fmt.Sprintf("%s%s?%s", "localhost", c.Request().URL.Path, c.QueryParams().Encode())
	return url
}

// Paginate :
func Paginate(value interface{}, pagination *Pagination, db *gorm.DB) func(db *gorm.DB) *gorm.DB {
	var totalRows int64
	db.Model(value).Count(&totalRows)

	pagination.TotalRows = totalRows
	pagination.TotalPages = int(math.Ceil(float64(totalRows) / float64(pagination.GetLimit())))

	return func(db *gorm.DB) *gorm.DB {
		return db.Offset(pagination.GetOffset()).Limit(pagination.GetLimit())
	}
}
