defmodule Architect.Projects.Branch do
  @enforce_keys [:name]
  defstruct [:name]

  @doc ~S"""
  Parses git `branch` stdout into list of Branch structs

  ## Examples

      iex> alias Architect.Projects.Branch
      ...> Branch.parse("  origin/HEAD -> origin/master\n  origin/git-changes\n  origin/master\n")
      [%Branch{name: "master"}, %Branch{name: "git-changes"}]

  """
  def parse({:ok, stdout}), do: parse(stdout)

  def parse({:error, error}), do: {:error, error}

  def parse(stdout) when is_binary(stdout) do
    stdout
    |> String.split("\n")
    |> Enum.map(&String.trim/1)
    |> Enum.reduce([], fn branch, acc ->
      case branch do
        "" ->
          acc

        "origin/HEAD" <> _ ->
          acc

        "origin/" <> name ->
          [%__MODULE__{name: name} | acc]
      end
    end)
  end

  @doc ~S"""
  Parses git `remote show origin` stdout into a single Branch struct

  ## Examples

      iex> alias Architect.Projects.Branch
      ...> Branch.parse_remote("* remote origin\n  Fetch URL: https://github.com/velocity-ci/velocity.git\n  Push  URL: https://github.com/velocity-ci/velocity.git\n  HEAD branch: master\n  Remote branches:\n    git-repository-changes tracked\n    master                 tracked\n    mobile-improvements    tracked\n    proof-of-concept       tracked\n  Local branch configured for 'git pull':\n    master merges with remote master\n  Local ref configured for 'git push':\n    master pushes to master (up to date)\n")
      %Branch{name: "master"}

  """
  def parse_remote({:ok, stdout}), do: parse_remote(stdout)

  def parse_remote({:error, error}), do: {:error, error}

  def parse_remote(stdout) when is_binary(stdout) do
    name =
      stdout
      |> String.split("\n")
      |> Enum.at(3)
      |> String.split(":")
      |> List.last()
      |> String.trim()

    %__MODULE__{name: name}
  end
end
