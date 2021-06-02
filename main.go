package main

import (
	"context"
	"database/sql"
	"entgo.io/ent/dialect"
	entsql "entgo.io/ent/dialect/sql"
	"fmt"
	"os"
	"todo/ent"
	"todo/ent/migrate"

	_ "github.com/jackc/pgx/v4/stdlib"
)

func checkerr(err error) {
	if err != nil {
		panic(err)
	}
}

func logger(next ent.Mutator) ent.Mutator {
	return ent.MutateFunc(func(ctx context.Context, m ent.Mutation) (ent.Value, error) {
		fmt.Println("Hook called")
		defer func() {
			var id int64
			if idr, ok := m.(interface{ ID() (int64, bool) }); ok {
				var exists bool

				id, exists = idr.ID()
				if !exists {
					fmt.Println("id not found")
				}
				fmt.Printf("Row ID : %v (%v)\n", id, exists)
			} else {
				panic("interface assertion not ok")
			}

			fmt.Println("Row ID through Field() :")
			fmt.Println(m.Field("id"))
		}()

		return next.Mutate(ctx, m)
	})
}

func main() {
	// init stuff
	db, err := sql.Open("pgx", os.Getenv("PG_DSN"))
	checkerr(err)
	checkerr(db.Ping())
	defer db.Close()

	pgClient := ent.NewClient(ent.Debug(), ent.Driver(entsql.OpenDB(dialect.Postgres, db)))
	defer pgClient.Close()

	checkerr(pgClient.Schema.Create(
		context.Background(),
		migrate.WithDropIndex(true),
		migrate.WithDropColumn(true),
	))

	pgClient.Use(logger)

	// trying out migrations
	fmt.Println("-> CREATE")
	obj, err := pgClient.Todo.Create().SetName("Test").Save(context.Background())
	checkerr(err)

	fmt.Println("New object created with ID :", obj.ID)
	fmt.Println("\n-> UPDATEONE")
	obj, err = pgClient.Todo.UpdateOneID(obj.ID).SetName("TestModified").Save(context.Background())
	checkerr(err)

	fmt.Println("Updated object with ID :", obj.ID)
}
