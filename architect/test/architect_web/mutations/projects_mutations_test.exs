defmodule ArchitectWeb.Mutations.ProjectMutationsTest do
  alias Architect.Projects
  use ArchitectWeb.ConnCase
  import Kronky.TestHelper

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
          createProject(address: \"https://github.com/velocity-ci/velocity.git\") {
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

      Projects.get_project_by_slug!("velocity-ci-velocity-at-github-com")
      |> assert_mutation_success(actual, @fields)
    end

    test "Failure - Unauthorized" do
      mutation = "
        mutation {
          createProject(address: \"https://github.com/velocity-ci/velocity.git\") {
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
