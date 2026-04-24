package mongo

import (
	"go/ast"
	"go/parser"
	"go/token"
	"os"
	"path/filepath"
	"testing"
)

func TestClientOptionsDisableRetryableWritesForStandaloneCompatibility(t *testing.T) {
	opts := clientOptions("mongodb://localhost:27017")
	if opts.RetryWrites == nil {
		t.Fatal("expected retryable writes to be set explicitly")
	}
	if *opts.RetryWrites {
		t.Fatal("expected retryable writes to be disabled")
	}
}

func TestRepositoriesAvoidMongoTransactionsForStandaloneCompatibility(t *testing.T) {
	repositories := []string{
		filepath.Join("..", "..", "comments", "repository.go"),
		filepath.Join("..", "..", "posts", "repository.go"),
		filepath.Join("..", "..", "users", "repository.go"),
	}

	for _, repository := range repositories {
		t.Run(repository, func(t *testing.T) {
			source, err := os.ReadFile(repository)
			if err != nil {
				t.Fatalf("read repository source: %v", err)
			}

			fileSet := token.NewFileSet()
			parsed, err := parser.ParseFile(fileSet, repository, source, 0)
			if err != nil {
				t.Fatalf("parse repository source: %v", err)
			}

			ast.Inspect(parsed, func(node ast.Node) bool {
				selector, ok := node.(*ast.SelectorExpr)
				if !ok {
					return true
				}

				switch selector.Sel.Name {
				case "StartSession", "WithTransaction":
					position := fileSet.Position(selector.Pos())
					t.Errorf("standalone MongoDB path must not call %s at %s", selector.Sel.Name, position)
				}

				return true
			})
		})
	}
}
