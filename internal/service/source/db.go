package source

import (
	"context"
	"errors"

	pb "github.com/heptaliane/katarive-go-sdk/gen/pb/plugin/v1"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/heptaliane/katarive-server/internal/model"
)

type DatabaseSourceRegistry struct {
	db  *gorm.DB
	sms []SourceManager
}

func (r *DatabaseSourceRegistry) SourceItem(
	ctx context.Context,
	url string,
) (*model.SourceItem, error) {
	sm, err := r.find(url)
	if err != nil {
		return nil, err
	}
	plugin := sm.Name()

	var item SourceItem
	err = r.db.First(&item, &SourceItem{
		Url:    url,
		Plugin: plugin,
	}).Error
	if err == nil {
		return item.IntoSourceItem(), nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	req := pb.GetSourceItemRequest{Url: url}
	res, err := sm.GetSourceItem(ctx, &req)
	if err != nil {
		return nil, err
	}

	item.FromSourceItem(res.GetItem())
	item.Plugin = plugin
	err = r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "Id"}, {Name: "Plugin"}},
		DoUpdates: item.Assignments(),
	}).Create(&item).Error

	return item.IntoSourceItem(), err
}
func (r *DatabaseSourceRegistry) SourceCollection(
	ctx context.Context,
	url string,
) (*model.SourceCollection, error) {
	collection, err := r.getOrCreateSourceCollection(ctx, url)
	if err != nil {
		return nil, err
	}

	return collection.IntoSourceCollection(), nil
}
func (r *DatabaseSourceRegistry) SourceItems(
	ctx context.Context,
	url string,
) ([]*model.SourceSummary, error) {
	collection, err := r.getOrCreateSourceCollection(ctx, url)
	if err != nil {
		return nil, err
	}

	return collection.IntoSourceItems(), nil
}

// helpers
func (r *DatabaseSourceRegistry) find(url string) (SourceManager, error) {
	for _, sm := range r.sms {
		if sm.IsSupported(url) {
			return sm, nil
		}
	}
	return nil, &model.UnsupportedSourceURLError{Url: url}
}
func (r *DatabaseSourceRegistry) getOrCreateSourceCollection(
	ctx context.Context,
	url string,
) (*SourceCollection, error) {
	sm, err := r.find(url)
	if err != nil {
		return nil, err
	}
	plugin := sm.Name()

	var collection SourceCollection
	err = r.db.
		Preload("Items").
		Preload("CollectionTags.Tag").
		First(&collection, &SourceCollection{
			Url:    url,
			Plugin: plugin,
		}).Error
	if err == nil {
		return &collection, nil
	}
	if !errors.Is(err, gorm.ErrRecordNotFound) {
		return nil, err
	}

	req := pb.GetSourceCollectionRequest{Url: url}
	res, err := sm.GetSourceCollection(ctx, &req)
	if err != nil {
		return nil, err
	}

	sc := res.GetCollection()
	collection.FromSourceCollection(sc)
	collection.Plugin = plugin
	err = r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "Id"}, {Name: "Plugin"}},
		UpdateAll: true,
	}).Create(&collection).Error

	var ctags []CollectionTag
	for _, t := range sc.GetTags() {
		var tag Tag
		err := r.db.Where(Tag{Name: t}).FirstOrCreate(&tag).Error
		if err != nil {
			return nil, err
		}

		var ctag CollectionTag
		err = r.db.Where(CollectionTag{
			CollectionId: sc.GetId(),
			Plugin:       plugin,
			TagId:        tag.Id,
		}).FirstOrCreate(&ctag).Error
		if err != nil {
			return nil, err
		}
		ctag.Tag = tag

		ctags = append(ctags, ctag)
	}

	var items []SourceItem
	for _, ss := range res.GetSources() {
		var item SourceItem
		item.FromSourceSummary(ss)
		item.Plugin = plugin
		item.CollectionId = &collection.Id
		err = r.db.Clauses(clause.OnConflict{
			Columns:   []clause.Column{{Name: "Id"}, {Name: "Plugin"}},
			DoUpdates: item.Assignments(),
		}).Create(&item).Error
		if err != nil {
			return nil, err
		}

		items = append(items, item)
	}

	collection.CollectionTags = ctags
	collection.Items = items

	return &collection, nil
}
func NewDatabaseSourceRegistry(
	db *gorm.DB,
	sms []SourceManager,
) *DatabaseSourceRegistry {
	db.AutoMigrate(
		&SourceCollection{},
		&SourceItem{},
		&Tag{},
		&CollectionTag{},
	)

	return &DatabaseSourceRegistry{
		db:  db,
		sms: sms,
	}
}

// GORM models
type SourceCollection struct {
	Id          string `gorm:"primaryKey"`
	Plugin      string `gorm:"primaryKey"`
	Url         string
	Title       string
	Description string
	Author      string

	CollectionTags []CollectionTag `gorm:"foreignKey:CollectionId,Plugin;references:Id,Plugin"`
	Items          []SourceItem    `gorm:"foreignKey:CollectionId,Plugin;references:Id,Plugin"`
}
type SourceItem struct {
	Id           string `gorm:"primaryKey"`
	Plugin       string `gorm:"primaryKey"`
	CollectionId *string
	Url          string
	Title        string
	Content      *string
	Language     int32
}
type Tag struct {
	Id   uint   `gorm:"primaryKey;autoIncrement"`
	Name string `gorm:"uniqueIndex;not null"`
}
type CollectionTag struct {
	Id           uint   `gorm:"primaryKey;autoIncrement"`
	CollectionId string `gorm:"index;not null"`
	Plugin       string `gorm:"index;not null"`
	TagId        uint   `gorm:"index;not null"`

	Tag Tag `gorm:"foreignKey:TagId"`
}

// model helpers
func (c *SourceCollection) FromSourceCollection(sc *model.SourceCollection) {
	c.Id = sc.GetId()
	c.Url = sc.GetUrl()
	c.Title = sc.GetTitle()
	c.Description = sc.GetDescription()
	c.Author = sc.GetAuthor()
}
func (c *SourceCollection) IntoSourceCollection() *model.SourceCollection {
	var tags []string
	for _, t := range c.CollectionTags {
		tags = append(tags, t.Tag.Name)
	}
	collection := &model.SourceCollection{
		Id:          c.Id,
		Url:         c.Url,
		Title:       c.Title,
		Description: c.Description,
		Author:      c.Author,
		Tags:        tags,
	}
	return collection
}
func (c *SourceCollection) IntoSourceItems() []*model.SourceSummary {
	var sources []*model.SourceSummary
	for _, i := range c.Items {
		sources = append(sources, i.IntoSourceSummary())
	}
	return sources
}
func (i *SourceItem) Assignments() clause.Set {
	fields := make(map[string]any)
	fields["id"] = i.Id
	fields["plugin"] = i.Plugin
	if i.CollectionId != nil {
		fields["collection_id"] = i.CollectionId
	}
	fields["url"] = i.Url
	fields["title"] = i.Title
	if i.Content != nil {
		fields["content"] = i.Content
	}
	fields["language"] = i.Language
	return clause.Assignments(fields)
}
func (i *SourceItem) FromSourceItem(si *model.SourceItem) {
	i.Id = si.GetId()
	i.Url = si.GetUrl()
	i.Title = si.GetTitle()
	i.Language = int32(si.GetLanguage())
	if collectionId := si.GetCollectionId(); collectionId != "" {
		i.CollectionId = &collectionId
	}
	if content := si.GetContent(); content != "" {
		i.Content = &content
	}
}
func (i *SourceItem) FromSourceSummary(ss *model.SourceSummary) {
	i.Id = ss.GetId()
	i.Url = ss.GetUrl()
	i.Title = ss.GetTitle()
}
func (i *SourceItem) IntoSourceItem() *model.SourceItem {
	item := &model.SourceItem{
		Id:           i.Id,
		CollectionId: i.CollectionId,
		Url:          i.Url,
		Title:        i.Title,
		Language:     pb.Language(i.Language),
	}
	if i.Content != nil {
		item.Content = *i.Content
	}
	return item
}
func (i *SourceItem) IntoSourceSummary() *model.SourceSummary {
	return &model.SourceSummary{
		Id:    i.Id,
		Url:   i.Url,
		Title: i.Title,
	}
}
