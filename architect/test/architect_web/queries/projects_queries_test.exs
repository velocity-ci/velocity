defmodule ArchitectWeb.Queries.ProjectsTest do
  alias Architect.Projects
  use ArchitectWeb.ConnCase
  import Kronky.TestHelper

  @projects_attrs [
    %{
      name: "Velocity CI",
      address: "https://github.com/velocity-ci/velocity.git"
    },
    %{
      name: "Elixir JSON diff",
      address: "https://github.com/EddyLane/elixir_json_diff.git"
    }
  ]

  @fields %{
    id: :string,
    name: :string,
    slug: :string,
    address: :string,
    inserted_at: :date,
    updated_at: :date
  }

  setup do
    projects =
      for project_attrs <- @projects_attrs do
        {:ok, project} = Projects.create_project(project_attrs)

        project
      end

    [conn: build_conn(), projects: projects]
  end

  test "gets a list of all projects", %{conn: conn, projects: projects} do
    query = """
      {
        projects {
          id,
          name,
          slug,
          address,
          insertedAt,
          updatedAt
        }
      }
    """

    %{"data" => %{"projects" => actual}} =
      conn
      |> post("/v2", %{query: query})
      |> json_response(200)

    assert_equivalent_graphql(projects, actual, @fields)
  end
end
