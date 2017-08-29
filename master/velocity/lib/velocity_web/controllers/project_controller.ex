defmodule VelocityWeb.ProjectController do
  @moduledoc "provides project management functions.\n"

  use VelocityWeb, :controller
  use Guardian.Phoenix.Controller
  alias Velocity.Project
  alias Velocity.ChangesetView
  alias Velocity.ProjectValidator

  def index(conn, _params) do
    render conn, "index.html"
  end

  @doc "Handles creating a project via POST /projects"
  @spec create(Conn, {}, {}, {}) :: nil
  def create(conn, project_params, _user, _claims) do
    with changeset <- Project.changeset(%Project{}, project_params),
         {:ok, changeset} <- ProjectValidator.create(changeset),
         {:ok, project} <- Project.create(changeset) do
      # |> put_resp_header("location", user_path(conn, :show, user))
      conn
      |> put_status(:created)
      |> render("show.json", project: project)
    else
      {:error, changeset} ->
        conn
        |> put_status(:unprocessable_entity)
        |> render(ChangesetView, "error.json", changeset: changeset)
    end
  end
end
