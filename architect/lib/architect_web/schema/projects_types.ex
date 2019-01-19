defmodule ArchitectWeb.Schema.ProjectsTypes do
  use Absinthe.Schema.Notation

  object :project do
    field(:id, non_null(:id))
    field(:name, non_null(:string))
    field(:slug, non_null(:string))
    field(:address, non_null(:string))
    field(:inserted_at, non_null(:naive_datetime))
    field(:updated_at, non_null(:naive_datetime))
  end

  object :commit do
    field(:sha, non_null(:string))
    field(:author, non_null(:commit_author))
  end

  object :commit_author do
    field(:date, non_null(:naive_datetime))
    field(:email, non_null(:string))
    field(:name, non_null(:string))
  end
end
