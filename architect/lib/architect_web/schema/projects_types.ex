defmodule ArchitectWeb.Schema.ProjectsTypes do
  use Absinthe.Schema.Notation
  use Absinthe.Relay.Schema.Notation, :modern

  alias ArchitectWeb.Resolvers.Projects
  alias ArchitectWeb.Middleware.SetProjectToContext

  node object(:project) do
    field(:name, non_null(:string))
    field(:slug, non_null(:string))
    field(:address, non_null(:string))
    field(:inserted_at, non_null(:naive_datetime))
    field(:updated_at, non_null(:naive_datetime))

    field :default_branch, non_null(:branch) do
      middleware(SetProjectToContext)
      resolve(&Projects.get_default_branch_for_project/3)
    end

    connection field(:branches, node_type: :branch) do
      middleware(SetProjectToContext)
      resolve(&Projects.list_branches_for_project/2)
    end
  end

  node object(:branch) do
    field(:name, non_null(:string))

    field :commit_amount, non_null(:integer) do
      resolve(&Projects.get_commit_count_for_branch/3)
    end

    connection field(:commits, node_type: :commit) do
      resolve(&Projects.list_commits/3)
    end

    field :blueprints, non_null(list_of(non_null(:blueprint))) do
      resolve(&Projects.list_blueprints_for_branch/3)
    end
  end

  node object(:commit) do
    field(:sha, non_null(:string))
    field(:author, non_null(:commit_author))
    field(:gpg_fingerprint, non_null(:string))
    field(:message, non_null(:string))

    field :branches, non_null(list_of(non_null(:branch))) do
      resolve(&Projects.list_branches_for_commit/3)
    end

    field :blueprints, non_null(list_of(non_null(:blueprint))) do
      resolve(&Projects.list_blueprints_for_commit/3)
    end
  end

  node object(:commit_author) do
    field(:date, non_null(:naive_datetime))
    field(:email, non_null(:string))
    field(:name, non_null(:string))
  end

  node object(:blueprint) do
    field(:name, non_null(:string))
    field(:description, :string)
  end
end
