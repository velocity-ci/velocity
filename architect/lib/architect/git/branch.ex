defmodule Architect.Git.Branch do
  alias Architect.Git.Repository

  @enforce_keys [:name]
  defstruct [:name]

  def get_all(repo) do
    repo = Repository.get_repo(repo)
    Porcelain.exec("git", ["fetch", "--prune"], dir: repo.directory)
    res = Porcelain.exec("git", ["branch", "--remote"], dir: repo.directory)
    branches = String.split(res.out, "\n")

    Enum.map(branches, fn x -> String.trim(x) end)
    |> Enum.filter(fn x -> x != "" end)
  end
end
