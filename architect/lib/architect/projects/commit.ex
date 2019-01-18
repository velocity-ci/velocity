defmodule(Architect.Projects.Commit.Author, do: defstruct([:email, :name, :date]))

defmodule Architect.Projects.Commit do
  @enforce_keys [:sha]
  defstruct [:sha, :author, :gpg_fingerprint, :message]

  @doc ~S"""

  The format passed to the Git CLI for a commit

    see: https://git-scm.com/docs/pretty-formats

    %H: commit hash
    %aI: author date, strict ISO 8601 format
    %aE: author email (respecting .mailmap, see git-shortlog[1] or git-blame[1])
    %aN: author name (respecting .mailmap, see git-shortlog[1] or git-blame[1])
    %GF: show the fingerprint of the key used to sign a signed commit
    %s: subject

  """
  def format(), do: '%H%n%aI%n%aE%n%aN%n%GF%n%s'

  @doc ~S"""
  Parses commit output into a single Commit struct

  ## Examples

      ...> alias Architect.Projects.Commit
      ...> alias Commit.Author
      ...> parsed = Commit.parse("4c2439630bbea0bcad61adc78b434cc804117090\n2019-01-17T17:39:27+00:00\nvj@vjpatel.me\nVJ Patel\n%GF\nadd parsing of git shell for commits and branches\n5697be45fa5cb5474a49f489c822e2d290693037\n2019-01-16T23:21:41+00:00\nvj@vjpatel.me\nVJ Patel\n%GF\nWip: added a few basic git funcs")
      ...> {:ok, dt_first, _} = DateTime.from_iso8601("2019-01-17T17:39:27+00:00")
      ...> {:ok, dt_second, _} = DateTime.from_iso8601("2019-01-16T23:21:41+00:00")
      ...> with [%Commit{author: %Author{date: ^dt_first, email: "vj@vjpatel.me", name: "VJ Patel"}, gpg_fingerprint: "%GF", message: "add parsing of git shell for commits and branches", sha: "4c2439630bbea0bcad61adc78b434cc804117090"}, %Commit{author: %Author{date: ^dt_second, email: "vj@vjpatel.me", name: "VJ Patel"},gpg_fingerprint: "%GF",message: "Wip: added a few basic git funcs", sha: "5697be45fa5cb5474a49f489c822e2d290693037"}] <- parsed, do: :passed
      :passed

  """
  def parse(stdout) when is_binary(stdout) do
    stdout
    |> String.split("\n")
    |> Enum.chunk_every(6)
    |> Enum.filter(fn x -> x != [""] end)
    |> Enum.map(&parse_commit_lines/1)
  end

  @doc ~S"""
  Parses commit output into a single Commit struct

  ## Examples

      iex> alias Architect.Projects.Commit
      ...> alias Commit.Author
      ...> parsed = Commit.parse_show("a51ab658cbb564bdf1990952b8eefb101c1aa823\n2019-01-17T19:05:16+00:00\nnaedin@gmail.com\nEddy Lane\n%GF\n[architect] WIP\n")
      ...> {:ok, dt, _} = DateTime.from_iso8601("2019-01-17T19:05:16+00:00")
      ...> with %Commit{author: %Author{date: ^dt, email: "naedin@gmail.com", name: "Eddy Lane" }, gpg_fingerprint: "%GF", message: "[architect] WIP", sha: "a51ab658cbb564bdf1990952b8eefb101c1aa823"} <- parsed, do: :passed
      :passed

  """
  def parse_show(stdout) when is_binary(stdout) do
    stdout
    |> String.split("\n")
    |> parse_commit_lines()
  end

  defp parse_commit_lines(l) do
    {:ok, dt, _} = DateTime.from_iso8601(Enum.at(l, 1))

    %__MODULE__{
      sha: Enum.at(l, 0),
      author: %__MODULE__.Author{
        date: dt,
        email: Enum.at(l, 2),
        name: Enum.at(l, 3)
      },
      gpg_fingerprint: if(Enum.at(l, 4) != "", do: Enum.at(l, 4), else: nil),
      message: Enum.at(l, 5)
    }
  end
end
