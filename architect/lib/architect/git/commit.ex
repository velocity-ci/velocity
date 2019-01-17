defmodule(Architect.Git.Commit.Author, do: defstruct([:email, :name, :date]))

defmodule Architect.Git.Commit do
  alias Architect.Git.Repository

  @enforce_keys [:sha]
  defstruct [:sha, :author, :gpg_fingerprint, :message]

  # see: https://git-scm.com/docs/pretty-formats
  # %H: commit hash
  # %aI: author date, strict ISO 8601 format
  # %aE: author email (respecting .mailmap, see git-shortlog[1] or git-blame[1])
  # %aN: author name (respecting .mailmap, see git-shortlog[1] or git-blame[1])
  # %GF: show the fingerprint of the key used to sign a signed commit
  # %s: subject
  @format '%H%n%aI%n%aE%n%aN%n%GF%n%s'

  def all_for_branch(repo, branch) do
    repo = Repository.get_repo(repo)
    Architect.Git.Repository.checkout(repo, branch)
    res = Porcelain.exec("git", ["log", "--format=#{@format}"], dir: repo.directory)

    String.split(res.out, "\n")
    |> Enum.chunk_every(6)
    |> Enum.filter(fn x -> x != [""] end)
    |> Enum.map(fn l -> parse_commit_lines(l) end)
  end

  def by_sha(repo, sha) do
    repo = Repository.get_repo(repo)
    res = Porcelain.exec("git", ["show", "-s", "--format=#{@format}", sha], dir: repo.directory)

    String.split(res.out, "\n")
    |> parse_commit_lines()
  end

  def head_of_branch(repo, branch) do
    repo = Repository.get_repo(repo)
    Architect.Git.Repository.checkout(repo, branch)
    res = Porcelain.exec("git", ["rev-parse", "HEAD"], dir: repo.directory)

    String.trim(res.out)
  end

  defp parse_commit_lines(l) do
    {:ok, dt, _} = DateTime.from_iso8601(Enum.at(l, 1))

    %Architect.Git.Commit{
      sha: Enum.at(l, 0),
      author: %Architect.Git.Commit.Author{
        date: dt,
        email: Enum.at(l, 2),
        name: Enum.at(l, 3)
      },
      gpg_fingerprint: if(Enum.at(l, 4) != "", do: Enum.at(l, 4), else: nil),
      message: Enum.at(l, 5)
    }
  end
end
