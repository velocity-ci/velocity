defmodule ArchitectWeb.V1.HealthView do
  use ArchitectWeb, :view

  def render("index.json", _) do
    %{
      data: %{
        status: "ok"
      }
    }
  end
end
