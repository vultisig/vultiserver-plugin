package postgres

import (
	"context"
	"encoding/json"
	"errors"
	"fmt"
	"reflect"
	"strings"

	"github.com/jackc/pgx/v5"
	"github.com/vultisig/vultisigner/common"
	"github.com/vultisig/vultisigner/internal/types"
)

const PLUGINS_TABLE = "plugins"
const PLUGIN_TAGS_TABLE = "plugin_tags"

func (P *PostgresBackend) collectPlugins(rows pgx.Rows) ([]types.Plugin, error) {
	defer rows.Close()

	var plugins []types.Plugin
	pluginMap := make(map[string]*types.Plugin)
	for rows.Next() {
		var plugin types.Plugin
		var tagId *string
		var tagName *string
		var tagColor *string

		err := rows.Scan(
			&plugin.ID,
			&plugin.CreatedAt,
			&plugin.UpdatedAt,
			&plugin.Type,
			&plugin.Title,
			&plugin.Description,
			&plugin.Metadata,
			&plugin.ServerEndpoint,
			&plugin.PricingID,
			&plugin.CategoryID,
			&tagId,
			&tagName,
			&tagColor,
		)
		if err != nil {
			return plugins, err
		}

		// add plugin if does not already exist in the map
		if _, exists := pluginMap[plugin.ID]; !exists {
			plugin.Tags = []types.Tag{}
			pluginMap[plugin.ID] = &plugin
		}

		// add tag to plugin tag list
		if tagId != nil {
			pluginMap[plugin.ID].Tags = append(pluginMap[plugin.ID].Tags, types.Tag{
				ID:    *tagId,
				Name:  *tagName,
				Color: *tagColor,
			})
		}
	}

	// convert map to list
	for _, p := range pluginMap {
		plugins = append(plugins, *p)
	}

	return plugins, nil
}

func (p *PostgresBackend) FindPluginById(ctx context.Context, id string) (*types.Plugin, error) {
	query := fmt.Sprintf(
		`SELECT p.*, t.*
		FROM %s p
		LEFT JOIN plugin_tags pt ON p.id = pt.plugin_id
		LEFT JOIN tags t ON pt.tag_id = t.id
		WHERE p.id = $1;`,
		PLUGINS_TABLE,
	)

	rows, err := p.pool.Query(ctx, query, id)
	if err != nil {
		return nil, err
	}

	plugins, err := p.collectPlugins(rows)
	if err != nil {
		return nil, err
	}

	plugin := plugins[0]

	return &plugin, nil
}

func (p *PostgresBackend) FindPlugins(
	ctx context.Context,
	filters types.PluginFilters,
	skip int,
	take int,
	sort string,
) (types.PluginsPaginatedList, error) {
	if p.pool == nil {
		return types.PluginsPaginatedList{}, fmt.Errorf("database pool is nil")
	}

	orderBy, orderDirection := common.GetSortingCondition(sort)

	joinQuery := fmt.Sprintf(`
		FROM %s p
		LEFT JOIN plugin_tags pt ON p.id = pt.plugin_id
		LEFT JOIN tags t ON pt.tag_id = t.id`,
		PLUGINS_TABLE,
	)

	query := `SELECT p.*, t.*` + joinQuery
	queryTotal := `SELECT COUNT(DISTINCT p.id) as total_count` + joinQuery

	args := []any{}
	argsTotal := []any{}
	currentArgNumber := 1

	// filters
	filterClause := "WHERE"
	if filters.Term != nil {
		queryFilter := fmt.Sprintf(
			` %s (p.title ILIKE $%d OR p.description ILIKE $%d)`,
			filterClause,
			currentArgNumber,
			currentArgNumber+1,
		)
		filterClause = "AND"
		currentArgNumber += 2

		term := "%" + *filters.Term + "%"
		args = append(args, term, term)
		argsTotal = append(argsTotal, term, term)

		query += queryFilter
		queryTotal += queryFilter
	}

	if filters.TagID != nil {
		queryFilter := fmt.Sprintf(
			` %s p.id IN (
				SELECT pti.plugin_id
    		FROM plugin_tags pti
    		JOIN tags ti ON pti.tag_id = ti.id
    		WHERE ti.id = $%d
			)`,
			filterClause,
			currentArgNumber,
		)

		queryFilterTotal := fmt.Sprintf(
			` %s t.id = $%d`,
			filterClause,
			currentArgNumber,
		)
		filterClause = "AND"
		currentArgNumber += 1

		args = append(args, filters.TagID)
		argsTotal = append(argsTotal, filters.TagID)

		query += queryFilter
		queryTotal += queryFilterTotal
	}

	if filters.CategoryID != nil {
		queryFilter := fmt.Sprintf(
			` %s p.category_id = $%d`,
			filterClause,
			currentArgNumber,
		)
		filterClause = "AND"
		currentArgNumber += 1

		args = append(args, filters.CategoryID)
		argsTotal = append(argsTotal, filters.CategoryID)

		query += queryFilter
		queryTotal += queryFilter
	}

	// pagination
	queryOrderPaginate := fmt.Sprintf(
		` ORDER BY p.%s %s LIMIT $%d OFFSET $%d;`,
		pgx.Identifier{orderBy}.Sanitize(),
		orderDirection,
		currentArgNumber,
		currentArgNumber+1,
	)
	args = append(args, take, skip)
	query += queryOrderPaginate

	queryTotal += " GROUP BY p.id;"

	// execute
	rows, err := p.pool.Query(ctx, query, args...)
	if err != nil {
		return types.PluginsPaginatedList{}, err
	}

	plugins, err := p.collectPlugins(rows)
	if err != nil {
		return types.PluginsPaginatedList{}, err
	}

	// execute total results count
	var totalCount int
	err = p.pool.QueryRow(ctx, queryTotal, argsTotal...).Scan(&totalCount)
	if err != nil {
		// exactly 1 row expected, if no results return empty list
		if err.Error() == "no rows in result set" {
			return types.PluginsPaginatedList{
				Plugins:    plugins,
				TotalCount: 0,
			}, nil
		}
		return types.PluginsPaginatedList{}, err
	}

	pluginsList := types.PluginsPaginatedList{
		Plugins:    plugins,
		TotalCount: totalCount,
	}

	return pluginsList, nil
}

func (p *PostgresBackend) CreatePlugin(ctx context.Context, pluginDto types.PluginCreateDto) (*types.Plugin, error) {
	query := fmt.Sprintf(`INSERT INTO %s (
		type,
		title,
		description,
		metadata,
		server_endpoint,
		pricing_id,
		category_id
	) VALUES (
		@Type,
		@Title,
		@Description,
		@Metadata,
		@ServerEndpoint,
		@PricingID,
		@CategoryID
	) RETURNING id;`, PLUGINS_TABLE)
	args := pgx.NamedArgs{
		"Type":           pluginDto.Type,
		"Title":          pluginDto.Title,
		"Description":    pluginDto.Description,
		"Metadata":       pluginDto.Metadata,
		"ServerEndpoint": pluginDto.ServerEndpoint,
		"PricingID":      pluginDto.PricingID,
		"CategoryID":     pluginDto.CategoryID,
	}

	var createdId string
	err := p.pool.QueryRow(ctx, query, args).Scan(&createdId)
	if err != nil {
		return nil, err
	}

	return p.FindPluginById(ctx, createdId)
}

func (p *PostgresBackend) UpdatePlugin(ctx context.Context, id string, updates types.PluginUpdateDto) (*types.Plugin, error) {
	t := reflect.TypeOf(updates)
	v := reflect.ValueOf(updates)
	numFields := t.NumField()

	query := fmt.Sprintf(`UPDATE %s SET `, PLUGINS_TABLE)
	args := pgx.NamedArgs{
		"id": id,
	}

	// iterate over dto props and assign non-empty for update
	var updateStatements []string
	for i := 0; i < numFields; i++ {
		field := t.Field(i) // field metadata
		value := v.Field(i) // field value

		// filter out json undefined values
		if !value.IsNil() {
			// get db field name (same as it is defined in the json input)
			fieldName := field.Tag.Get("json")
			if fieldName == "" {
				// fallback to prop name
				fieldName = field.Name
			}

			// get value from dto reference
			var fieldValue interface{}
			if field.Type == reflect.TypeOf((*json.RawMessage)(nil)) {
				// keep as reference to []byte
				fieldValue = value.Interface().(*json.RawMessage)
			} else {
				// dereference
				fieldValue = value.Elem().Interface()
			}

			updateStatements = append(updateStatements, fmt.Sprintf("%s = @%s", fieldName, fieldName))
			args[fieldName] = fieldValue
		}
	}

	if len(updateStatements) == 0 {
		return nil, errors.New("No updates provided")
	}

	query += strings.Join(updateStatements, ", ")
	query += " WHERE id = @id;"

	_, err := p.pool.Exec(ctx, query, args)
	if err != nil {
		return nil, err
	}

	return p.FindPluginById(ctx, id)
}

func (p *PostgresBackend) DeletePluginById(ctx context.Context, id string) error {
	query := fmt.Sprintf(`DELETE FROM %s WHERE id = $1;`, PLUGINS_TABLE)

	_, err := p.pool.Exec(ctx, query, id)
	if err != nil {
		return err
	}

	return nil
}

func (p *PostgresBackend) AttachTagToPlugin(ctx context.Context, pluginId string, tagId string) (*types.Plugin, error) {
	query := fmt.Sprintf(`INSERT INTO %s (
		plugin_id,
		tag_id
	) VALUES (
		@PluginID,
		@TagID
	);`, PLUGIN_TAGS_TABLE)
	args := pgx.NamedArgs{
		"PluginID": pluginId,
		"TagID":    tagId,
	}

	_, err := p.pool.Exec(ctx, query, args)
	if err != nil {
		return nil, err
	}

	return p.FindPluginById(ctx, pluginId)
}

func (p *PostgresBackend) DetachTagFromPlugin(ctx context.Context, pluginId string, tagId string) (*types.Plugin, error) {
	query := fmt.Sprintf(`DELETE FROM %s WHERE plugin_id = @PluginID AND tag_id = @TagID;`, PLUGIN_TAGS_TABLE)
	args := pgx.NamedArgs{
		"PluginID": pluginId,
		"TagID":    tagId,
	}

	_, err := p.pool.Exec(ctx, query, args)
	if err != nil {
		return nil, err
	}

	return p.FindPluginById(ctx, pluginId)
}
