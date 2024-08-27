package genconfig

import (
	"fmt"
	"gorm.io/gen"
	"gorm.io/gen/field"
	"gorm.io/gen/internal/generate"
)

type Config struct {
	items map[string]ModelConfig
}
type ModelConfig struct {
	TableName     string
	ModelName     string
	Relationships []Relationship
}
type Relationship struct {
	Type           field.RelationshipType
	ModelFieldName string
	Target         RelateTarget
	Join           *RelateTarget
}
type RelateTarget struct {
	Table      string
	ForeignKey string
	References string
}

type JoinTable struct {
	Model     string
	Filed     string
	JoinModel string
}

func NewConfig(modelConfigs ...ModelConfig) *Config {
	s := &Config{items: make(map[string]ModelConfig)}
	for _, conf := range modelConfigs {
		s.Add(conf)
	}

	return s
}

func (c *Config) Add(conf ModelConfig) {
	c.items[conf.TableName] = conf
}

func (c *Config) Build(g *gen.Generator) (modelNames []string, joinTableConfig []JoinTable, err error) {
	var queryStructMetas = make(map[string]*generate.QueryStructMeta)
	for tableName, config := range c.items {
		queryStructMetas[tableName] = g.GenerateModelAs(tableName, config.ModelName)
		modelNames = append(modelNames, queryStructMetas[tableName].ModelStructName)
	}
	for tableName, config := range c.items {
		if config.Relationships != nil {
			var opts []gen.ModelOpt
			for _, relationship := range config.Relationships {
				targetMeta, ok := queryStructMetas[relationship.Target.Table]
				if !ok {
					err = fmt.Errorf("relation table %s not defined", relationship.Target.Table)
					return
				}
				if relationship.Type != field.Many2Many {
					opts = append(opts, gen.FieldRelate(
						relationship.Type,
						relationship.ModelFieldName,
						targetMeta,
						foreignKeyConfig(relationship.Target.ForeignKey, relationship.Target.References),
					))
				} else {
					if relationship.Join == nil {
						err = fmt.Errorf("the table %s relation join config not defined", config.TableName)
						return
					}
					joinMeta, ok1 := queryStructMetas[relationship.Join.Table]
					if ok1 {
						joinTableConfig = append(joinTableConfig, JoinTable{queryStructMetas[tableName].ModelStructName, relationship.ModelFieldName, joinMeta.ModelStructName})
					}
					opts = append(opts, gen.FieldRelate(
						relationship.Type,
						relationship.ModelFieldName,
						targetMeta,
						many2ManyConfig(relationship.Target.ForeignKey, relationship.Join.ForeignKey, relationship.Join.Table, relationship.Join.References, relationship.Target.References),
					))
				}
			}
			queryStructMetas[tableName] = g.GenerateModelAs(
				tableName,
				config.ModelName,
				opts...,
			)
		}
	}
	var metas []interface{}
	for _, meta := range queryStructMetas {
		metas = append(metas, meta)
	}
	g.ApplyBasic(metas...)

	return
}

// ForeignKeyConfig
// hasOne    A.id   <--> B.A_id  foreignKey(B.A_id) reference (A.id)
// HasMany   A.id   <--> B.A_id  foreignKey(B.A_id) reference (A.id)
// belongTo  A.C_id <--> C.id    foreignKey(A.C_id) reference (C.id)
func foreignKeyConfig(foreignKey, references string) *field.RelateConfig {
	relateTag := field.NewGormTag()
	relateTag.Set("foreignKey", foreignKey)
	relateTag.Set("references", references)
	return &field.RelateConfig{GORMTag: relateTag}
}

// Many2ManyConfig A <--> join(B.A_id,B.C_id) <--> C   foreignKey(A.C_id) reference (C.id)
func many2ManyConfig(foreignKey, joinForeignKey, joinTable, JoinReferences, references string) *field.RelateConfig {
	relateTag := field.NewGormTag()
	relateTag.Set("many2many", joinTable)
	relateTag.Set("foreignKey", foreignKey)
	relateTag.Set("joinForeignKey", joinForeignKey)
	relateTag.Set("references", references)
	relateTag.Set("JoinReferences", JoinReferences)
	return &field.RelateConfig{GORMTag: relateTag}
}
