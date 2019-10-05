package main

import (
	"roman_group/models"
)

type NewsPaginator struct {
	NewsCounter int64
	NewsPerPage int64
	TotalPage   int64
	CurrentPage int64
	HasNext     bool
	HasPrev     bool
	PrevPage    int64
	NextPage    int64
	PageList    []int64
}

type ProjectPaginator struct {
	ProjectCounter int64
	ProjectPerPage int64
	TotalPage      int64
	CurrentPage    int64
	HasNext        bool
	HasPrev        bool
	PrevPage       int64
	NextPage       int64
	PageList       []int64
}

type FeedBackPaginator struct {
	FeedBackCounter int64
	FeedBackPerPage int64
	TotalPage       int64
	CurrentPage     int64
	HasNext         bool
	HasPrev         bool
	PrevPage        int64
	NextPage        int64
	PageList        []int64
}

func initNewsPaginator(page int64, perPage int64) *NewsPaginator {
	var np NewsPaginator
	np.NewsCounter = models.GetCountNews(db)
	np.NewsPerPage = perPage
	if np.NewsCounter%np.NewsPerPage != 0 {
		np.TotalPage = np.NewsCounter/np.NewsPerPage + 1
	} else {
		np.TotalPage = np.NewsCounter / np.NewsPerPage
	}
	np.CurrentPage = page
	if np.HasNext = np.hasNext(); np.HasNext {
		np.NextPage = np.CurrentPage + 1
	}
	if np.HasPrev = np.hasPrev(); np.HasPrev {
		np.PrevPage = np.CurrentPage - 1
	}

	var i int64

	if np.TotalPage > 8 {
		for i = 1; i <= np.TotalPage; i++ {
			if i < np.CurrentPage-2 || (i == 1 && np.CurrentPage > 2) || i > np.CurrentPage+2 || (i == np.TotalPage && np.CurrentPage < np.TotalPage-2) {
				continue
			} else {
				np.PageList = append(np.PageList, i)
			}
		}
	} else {
		for i = 1; i <= np.TotalPage; i++ {
			np.PageList = append(np.PageList, i)
		}
	}

	return &np
}

func (np *NewsPaginator) hasNext() bool {
	return np.TotalPage > np.CurrentPage
}

func (np *NewsPaginator) hasPrev() bool {
	return 1 < np.CurrentPage
}

func initProjectPaginator(page int64, perPage int64) *ProjectPaginator {
	var pp ProjectPaginator
	pp.ProjectCounter = models.GetCountProject(db)
	pp.ProjectPerPage = perPage
	if pp.ProjectCounter%pp.ProjectPerPage != 0 {
		pp.TotalPage = pp.ProjectCounter/pp.ProjectPerPage + 1
	} else {
		pp.TotalPage = pp.ProjectCounter / pp.ProjectPerPage
	}
	pp.CurrentPage = page
	if pp.HasNext = pp.hasNext(); pp.HasNext {
		pp.NextPage = pp.CurrentPage + 1
	}
	if pp.HasPrev = pp.hasPrev(); pp.HasPrev {
		pp.PrevPage = pp.CurrentPage - 1
	}

	var i int64

	if pp.TotalPage > 8 {
		for i = 1; i <= pp.TotalPage; i++ {
			if i < pp.CurrentPage-2 || (i == 1 && pp.CurrentPage > 2) || i > pp.CurrentPage+2 || (i == pp.TotalPage && pp.CurrentPage < pp.TotalPage-2) {
				continue
			} else {
				pp.PageList = append(pp.PageList, i)
			}
		}
	} else {
		for i = 1; i <= pp.TotalPage; i++ {
			pp.PageList = append(pp.PageList, i)
		}
	}

	return &pp
}

func (pp *ProjectPaginator) hasNext() bool {
	return pp.TotalPage > pp.CurrentPage
}

func (pp *ProjectPaginator) hasPrev() bool {
	return 1 < pp.CurrentPage
}

func iniFeedBackPaginator(page int64, perPage int64, checkStatus bool) *FeedBackPaginator {
	var fbp FeedBackPaginator
	fbp.FeedBackCounter = models.GetCountFeedBacks(db, checkStatus)
	fbp.FeedBackPerPage = perPage
	if fbp.FeedBackCounter%fbp.FeedBackPerPage != 0 {
		fbp.TotalPage = fbp.FeedBackCounter/fbp.FeedBackPerPage + 1
	} else {
		fbp.TotalPage = fbp.FeedBackCounter / fbp.FeedBackPerPage
	}
	fbp.CurrentPage = page
	if fbp.HasNext = fbp.hasNext(); fbp.HasNext {
		fbp.NextPage = fbp.CurrentPage + 1
	}
	if fbp.HasPrev = fbp.hasPrev(); fbp.HasPrev {
		fbp.PrevPage = fbp.CurrentPage - 1
	}

	var i int64

	if fbp.TotalPage > 8 {
		for i = 1; i <= fbp.TotalPage; i++ {
			if i < fbp.CurrentPage-2 || (i == 1 && fbp.CurrentPage > 2) || i > fbp.CurrentPage+2 || (i == fbp.TotalPage && fbp.CurrentPage < fbp.TotalPage-2) {
				continue
			} else {
				fbp.PageList = append(fbp.PageList, i)
			}
		}
	} else {
		for i = 1; i <= fbp.TotalPage; i++ {
			fbp.PageList = append(fbp.PageList, i)
		}
	}

	return &fbp
}

func (fbp *FeedBackPaginator) hasNext() bool {
	return fbp.TotalPage > fbp.CurrentPage
}

func (fbp *FeedBackPaginator) hasPrev() bool {
	return 1 < fbp.CurrentPage
}
