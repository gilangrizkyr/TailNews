package usecase

import (
	"context"

	"news-portal/apps/api/internal/domain/author"
	"news-portal/apps/api/internal/domain/category"
	"news-portal/apps/api/internal/domain/tag"
	"news-portal/apps/api/internal/pkg/slug"
	"news-portal/apps/api/internal/transport/http/dto"
)

type TaxonomyAdminUsecase struct {
	categoryRepo category.Repository
	tagRepo      tag.Repository
	authorRepo   author.Repository
}

func NewTaxonomyAdminUsecase(
	catRepo category.Repository,
	tagRepo tag.Repository,
	authorRepo author.Repository,
) *TaxonomyAdminUsecase {
	return &TaxonomyAdminUsecase{
		categoryRepo: catRepo,
		tagRepo:      tagRepo,
		authorRepo:   authorRepo,
	}
}

// Categories
func (u *TaxonomyAdminUsecase) ListCategories(ctx context.Context) ([]dto.CategoryResponse, error) {
	cats, err := u.categoryRepo.List(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]dto.CategoryResponse, len(cats))
	for i, c := range cats {
		res[i] = dto.CategoryResponse{
			ID:          c.ID,
			Name:        c.Name,
			Slug:        c.Slug,
			Description: c.Description,
			ParentID:    c.ParentID,
			SortOrder:   c.SortOrder,
			IsNav:       c.IsNav,
			CreatedAt:   c.CreatedAt,
			UpdatedAt:   c.UpdatedAt,
		}
	}
	return res, nil
}

func (u *TaxonomyAdminUsecase) CreateCategory(ctx context.Context, req *dto.CategoryRequest) (*dto.CategoryResponse, error) {
	catSlug := req.Slug
	if catSlug == "" {
		catSlug = slug.Generate(req.Name)
	}

	c := &category.Category{
		Name:        req.Name,
		Slug:        catSlug,
		Description: req.Description,
		ParentID:    req.ParentID,
		SortOrder:   req.SortOrder,
		IsNav:       req.IsNav,
	}

	id, err := u.categoryRepo.Create(ctx, c)
	if err != nil {
		return nil, err
	}
	c.ID = id

	return &dto.CategoryResponse{
		ID:          c.ID,
		Name:        c.Name,
		Slug:        c.Slug,
		Description: c.Description,
		ParentID:    c.ParentID,
		SortOrder:   c.SortOrder,
		IsNav:       c.IsNav,
		CreatedAt:   c.CreatedAt,
		UpdatedAt:   c.UpdatedAt,
	}, nil
}

func (u *TaxonomyAdminUsecase) UpdateCategory(ctx context.Context, id uint64, req *dto.CategoryRequest) error {
	catSlug := req.Slug
	if catSlug == "" {
		catSlug = slug.Generate(req.Name)
	}

	c := &category.Category{
		ID:          id,
		Name:        req.Name,
		Slug:        catSlug,
		Description: req.Description,
		ParentID:    req.ParentID,
		SortOrder:   req.SortOrder,
		IsNav:       req.IsNav,
	}

	return u.categoryRepo.Update(ctx, c)
}

func (u *TaxonomyAdminUsecase) DeleteCategory(ctx context.Context, id uint64) error {
	return u.categoryRepo.Delete(ctx, id)
}

// Tags
func (u *TaxonomyAdminUsecase) ListTags(ctx context.Context) ([]dto.TagResponse, error) {
	tags, err := u.tagRepo.List(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]dto.TagResponse, len(tags))
	for i, t := range tags {
		res[i] = dto.TagResponse{
			ID:        t.ID,
			Name:      t.Name,
			Slug:      t.Slug,
			CreatedAt: t.CreatedAt,
			UpdatedAt: t.UpdatedAt,
		}
	}
	return res, nil
}

func (u *TaxonomyAdminUsecase) CreateTag(ctx context.Context, req *dto.TagRequest) (*dto.TagResponse, error) {
	tSlug := req.Slug
	if tSlug == "" {
		tSlug = slug.Generate(req.Name)
	}

	t := &tag.Tag{
		Name: req.Name,
		Slug: tSlug,
	}

	id, err := u.tagRepo.Create(ctx, t)
	if err != nil {
		return nil, err
	}
	t.ID = id

	return &dto.TagResponse{
		ID:        t.ID,
		Name:      t.Name,
		Slug:      t.Slug,
		CreatedAt: t.CreatedAt,
		UpdatedAt: t.UpdatedAt,
	}, nil
}

func (u *TaxonomyAdminUsecase) DeleteTag(ctx context.Context, id uint64) error {
	return u.tagRepo.Delete(ctx, id)
}

// Authors
func (u *TaxonomyAdminUsecase) ListAuthors(ctx context.Context) ([]dto.AuthorResponse, error) {
	authors, err := u.authorRepo.List(ctx)
	if err != nil {
		return nil, err
	}
	res := make([]dto.AuthorResponse, len(authors))
	for i, a := range authors {
		res[i] = dto.AuthorResponse{
			ID:            a.ID,
			UserID:        a.UserID,
			Name:          a.Name,
			Slug:          a.Slug,
			Bio:           a.Bio,
			AvatarMediaID: a.AvatarMediaID,
			SocialLinks:   a.SocialLinks,
			CreatedAt:     a.CreatedAt,
			UpdatedAt:     a.UpdatedAt,
		}
	}
	return res, nil
}

func (u *TaxonomyAdminUsecase) CreateAuthor(ctx context.Context, req *dto.AuthorRequest) (*dto.AuthorResponse, error) {
	aSlug := req.Slug
	if aSlug == "" {
		aSlug = slug.Generate(req.Name)
	}

	a := &author.Author{
		UserID:        req.UserID,
		Name:          req.Name,
		Slug:          aSlug,
		Bio:           req.Bio,
		AvatarMediaID: req.AvatarMediaID,
		SocialLinks:   req.SocialLinks,
	}

	id, err := u.authorRepo.Create(ctx, a)
	if err != nil {
		return nil, err
	}
	a.ID = id

	return &dto.AuthorResponse{
		ID:            a.ID,
		UserID:        a.UserID,
		Name:          a.Name,
		Slug:          a.Slug,
		Bio:           a.Bio,
		AvatarMediaID: a.AvatarMediaID,
		SocialLinks:   a.SocialLinks,
		CreatedAt:     a.CreatedAt,
		UpdatedAt:     a.UpdatedAt,
	}, nil
}

func (u *TaxonomyAdminUsecase) UpdateAuthor(ctx context.Context, id uint64, req *dto.AuthorRequest) error {
	aSlug := req.Slug
	if aSlug == "" {
		aSlug = slug.Generate(req.Name)
	}

	a := &author.Author{
		ID:            id,
		UserID:        req.UserID,
		Name:          req.Name,
		Slug:          aSlug,
		Bio:           req.Bio,
		AvatarMediaID: req.AvatarMediaID,
		SocialLinks:   req.SocialLinks,
	}

	return u.authorRepo.Update(ctx, a)
}
