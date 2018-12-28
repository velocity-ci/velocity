defmodule ArchitectWeb.Schema.ProjectsTypes do
  use Absinthe.Schema.Notation

  object :project do
    field(:id, :id)
    field(:name, non_null(:string))
    field(:slug, non_null(:string))
    field(:repository, non_null(:string))
  end
end
