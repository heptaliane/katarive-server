package source

import (
	"context"
	"errors"
	"fmt"
	"strings"

	pb "github.com/heptaliane/katarive-go-sdk/gen/pb/plugin/v1"
	"gorm.io/gorm"
	"gorm.io/gorm/clause"

	"github.com/heptaliane/katarive-server/internal/model"
)

type DatabaseSourceRegistry struct {
	db  *gorm.DB
	sms []SourceManager
}

func (r *DatabaseSourceRegistry) AddItem(
	ctx context.Context,
	itemUrl string,
) error {
	sm, err := r.findItem(itemUrl)
	if err != nil {
		return err
	}

	// Fetch source
	req := pb.GetSourceItemRequest{Url: itemUrl}
	res, err := sm.GetSourceItem(ctx, &req)
	if err != nil {
		return err
	}

	// Add registry
	var item SourceItem
	item.FromSourceItem(res.GetItem())
	item.Plugin = sm.Name()
	return r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "Id"}, {Name: "Plugin"}},
		DoUpdates: item.Assignments(),
	}).Create(&item).Error
}
func (r *DatabaseSourceRegistry) AddCollection(
	ctx context.Context,
	collectionUrl string,
) error {
	sm, err := r.findCollection(collectionUrl)
	if err != nil {
		return err
	}

	// Fetch source
	req := pb.GetSourceCollectionRequest{Url: collectionUrl}
	res, err := sm.GetSourceCollection(ctx, &req)
	if err != nil {
		return err
	}

	// Add registry
	sc := res.GetCollection()
	plugin := sm.Name()

	var collection SourceCollection
	collection.FromSourceCollection(sc)
	collection.Plugin = plugin
	err = r.db.Clauses(clause.OnConflict{
		Columns:   []clause.Column{{Name: "Id"}, {Name: "Plugin"}},
		UpdateAll: true,
	}).Create(&collection).Error
	if err != nil {
		return err
	}

	for _, t := range sc.GetTags() {
		var tag Tag
		err := r.db.Where(Tag{Name: t}).FirstOrCreate(&tag).Error
		if err != nil {
			return err
		}

		var ctag CollectionTag
		err = r.db.Where(CollectionTag{
			CollectionId: sc.GetId(),
			Plugin:       plugin,
			TagId:        tag.Id,
		}).FirstOrCreate(&ctag).Error
		if err != nil {
			return err
		}
		ctag.Tag = tag
	}

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
			return err
		}
	}

	return nil
}
func (r *DatabaseSourceRegistry) GetItem(itemUrl string) (*model.SourceItem, error) {
	sm, err := r.findItem(itemUrl)
	if err != nil {
		return nil, err
	}

	var item SourceItem
	err = r.db.First(&item, &SourceItem{
		Url:    itemUrl,
		Plugin: sm.Name(),
	}).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	if item.Content == nil {
		return nil, nil
	}
	return item.IntoSourceItem(), nil
}
func (r *DatabaseSourceRegistry) GetItems(
	opts ...GetSourceOption,
) ([]*model.SourceSummary, error) {
	var options getSourceOptions
	for _, opt := range opts {
		opt(&options)
	}

	var items []*SourceItem
	err := r.db.
		Joins("Collection").
		Scopes(
			filterByCollection("Collection", &options),
			filterByItem("source_items", &options),
		).
		Find(&items).Debug().Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []*model.SourceSummary{}, nil
		}
		return nil, err
	}

	var mitems []*model.SourceSummary
	for _, item := range items {
		mitems = append(mitems, item.IntoSourceSummary())
	}
	return mitems, nil
}
func (r *DatabaseSourceRegistry) GetCollection(
	collectionUrl string,
) (*model.SourceCollection, error) {
	sm, err := r.findCollection(collectionUrl)
	if err != nil {
		return nil, err
	}

	var collection SourceCollection
	err = r.db.Preload("CollectionTags.Tag").
		First(&collection, &SourceCollection{
			Url:    collectionUrl,
			Plugin: sm.Name(),
		}).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return nil, nil
		}
		return nil, err
	}
	return collection.IntoSourceCollection(), nil
}
func (r *DatabaseSourceRegistry) GetCollections(
	opts ...GetSourceOption,
) ([]*model.SourceCollection, error) {
	var options getSourceOptions
	for _, opt := range opts {
		opt(&options)
	}

	var collections []*SourceCollection
	err := r.db.
		Preload("CollectionTags.Tag").
		Joins(joinStatement(
			"INNER JOIN",
			"source_collections", "source_items",
			[]string{"id", "plugin"}, []string{"collection_id", "plugin"},
		)).
		Scopes(
			filterByCollection("source_collections", &options),
			filterByItem("source_items", &options),
		).
		Distinct("source_collections.*").
		Find(&collections).Error
	if err != nil {
		if errors.Is(err, gorm.ErrRecordNotFound) {
			return []*model.SourceCollection{}, nil
		}
		return nil, err
	}

	var cs []*model.SourceCollection
	for _, collection := range collections {
		cs = append(cs, collection.IntoSourceCollection())
	}
	return cs, nil
}

// Ensure SourceRegistry implementation
var _ SourceRegistry = new(DatabaseSourceRegistry)

// helpers
func (r *DatabaseSourceRegistry) findItem(url string) (SourceManager, error) {
	for _, sm := range r.sms {
		if sm.IsSupportedItem(url) {
			return sm, nil
		}
	}
	return nil, &model.UnsupportedSourceURLError{Url: url}
}
func (r *DatabaseSourceRegistry) findCollection(url string) (SourceManager, error) {
	for _, sm := range r.sms {
		if sm.IsSupportedCollection(url) {
			return sm, nil
		}
	}
	return nil, &model.UnsupportedSourceURLError{Url: url}
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

	Collection *SourceCollection `gorm:"foreignKey:CollectionId,Plugin;references:Id,Plugin"`
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
	if i.Language != 0 {
		fields["language"] = i.Language
	}
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

// getSourceOptions helpers
func filterByCollection(table string, options *getSourceOptions) func(tx *gorm.DB) *gorm.DB {
	return func(tx *gorm.DB) *gorm.DB {
		if options.collectionUrl != "" {
			tx = tx.Where(fmt.Sprintf("`%s`.url = ?", table), options.collectionUrl)
		}
		return tx
	}
}
func filterByItem(table string, options *getSourceOptions) func(tx *gorm.DB) *gorm.DB {
	return func(tx *gorm.DB) *gorm.DB {
		if options.itemUrl != "" {
			tx = tx.Where(fmt.Sprintf("`%s`.url = ?", table), options.itemUrl)
		}
		return tx
	}
}
func joinStatement(joinType, t1, t2 string, fs1, fs2 []string) string {
	var conds []string
	for i := range fs1 {
		conds = append(
			conds,
			fmt.Sprintf("%s.%s = %s.%s", t1, fs1[i], t2, fs2[i]),
		)
	}
	return fmt.Sprintf("%s %s ON %s", joinType, t2, strings.Join(conds, " AND "))
}
