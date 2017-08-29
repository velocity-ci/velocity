defmodule VelocityWeb.ProjectView do
  @moduledoc "provides project output\n"

  use VelocityWeb, :view

  @spec render(String, {}) :: {}
  def render("index.json", %{projects: projects}) do
    %{data: render_many(projects, VelocityWeb.ProjectView, "project.json")}
  end

  @spec render(String, {}) :: {}
  def render("show.json", %{project: project}) do
    %{data: render_one(project, VelocityWeb.ProjectView, "project.json")}
  end

  @spec render(String, {:project}) :: {}
  def render("project.json", %{project: project}) do
    %{data:
      %{
        id: project.id_name,
        name: project.name
      }
    }
  end
end
