defmodule ArchitectWeb.Schema.ProjectsTypes do
  use Absinthe.Schema.Notation

  object :project do
    field(:id, non_null(:id))
    field(:name, non_null(:string))
    field(:slug, non_null(:string))
    field(:repository, non_null(:string))
    field(:inserted_at, non_null(:naive_datetime))
    field(:updated_at, non_null(:naive_datetime))
  end
end
