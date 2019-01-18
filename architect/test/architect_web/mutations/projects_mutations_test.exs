defmodule ArchitectWeb.Mutations.ProjectMutationsTest do
  alias Architect.Projects
  alias Architect.Projects.Project
  use ArchitectWeb.ConnCase
  import Kronky.TestHelper
  alias Kronky.ValidationMessage
  alias Architect.Repo

  @fields %{
    id: :string,
    name: :string,
    slug: :string,
    address: :string,
    inserted_at: :date,
    updated_at: :date
  }

  setup do
    []
  end

  describe "createProject" do
    test "Success" do
      mutation = "
        mutation {
          createProject(name: \"Velocity\", address: \"https://github.com/velocity-ci/velocity.git\") {
            result {
              id,
              name,
              slug,
              address,
              insertedAt,
              updatedAt
            },
            successful
          }
        }
      "

      %{"createProject" => actual} =
        graphql_request(mutation)
        |> expect_success!()

      Projects.get_project_by_slug!("velocity")
      |> assert_mutation_success(actual, @fields)
    end

    test "Failure - Unauthorized" do
      mutation = "
        mutation {
          createProject(name: \"Velocity\", address: \"https://github.com/velocity-ci/velocity.git\") {
            result {
              id
            }
          }
        }
      "

      unauthorized_graphql_request(mutation)
      |> expect_failure!()
      |> Enum.find(fn %{"message" => message} -> message == "Unauthorized" end)
      |> assert
    end
  end
end
