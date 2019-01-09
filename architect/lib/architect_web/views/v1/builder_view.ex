defmodule ArchitectWeb.V1.BuilderView do
  use ArchitectWeb, :view
  alias ArchitectWeb.V1.BuilderView

  def render("index.json", %{builders: builders}) do
    %{data: render_many(builders, BuilderView, "builder.json")}
  end

  def render("show.json", %{builder: builder}) do
    %{data: render_one(builder, BuilderView, "builder.json")}
  end

  def render("builder.json", %{builder: builder}) do
    %{
      id: builder.id,
      token: builder.token
    }
  end
end
