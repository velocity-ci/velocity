defmodule ArchitectWeb.Schema.ProjectsTypes do
  use Absinthe.Schema.Notation
  import Absinthe.Resolution.Helpers
  alias Architect.Projects
  alias Architect.Projects.Project

  object :project do
    field(:id, non_null(:id))
    field(:name, non_null(:string))
    field(:slug, non_null(:string))
    field(:address, non_null(:string))
    field(:inserted_at, non_null(:naive_datetime))
    field(:updated_at, non_null(:naive_datetime))

    field :branches, list_of(:branch) do
      resolve(fn %Project{} = project, _args, _resolution ->
        {:ok, Projects.list_branches(project)}
      end)
    end
  end

  object :commit do
    field(:sha, non_null(:string))
    field(:author, non_null(:commit_author))
    field(:gpg_fingerprint, non_null(:string))
    field(:message, non_null(:string))
  end

  object :commit_author do
    field(:date, non_null(:naive_datetime))
    field(:email, non_null(:string))
    field(:name, non_null(:string))
  end

  object :branch do
    field(:name, non_null(:string))
  end
end
