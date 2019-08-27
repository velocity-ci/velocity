defmodule Git.Commit do
  @keys [:sha, :author, :gpg_fingerprint, :message]

  @enforce_keys @keys
  defstruct @keys

  defmodule(Author, do: defstruct([:email, :name, :date]))

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
  def format(), do: "%H%n%aI%n%aE%n%aN%n%GF%n%s"

  @doc ~S"""
  Parses commit output into a list of Commit structs
  """
  def parse({:ok, stdout}), do: parse(stdout)

  def parse({:error, error}), do: {:error, error}

  def parse(stdout) when is_binary(stdout) do
    stdout
    |> String.split("\n")
    |> Enum.chunk_every(6)
    |> Enum.filter(fn x -> x != [""] end)
    |> Enum.map(&parse_commit_lines/1)
  end

  @doc ~S"""
  Parses commit output into a single Commit struct
  """
  def parse_show({:ok, stdout}), do: parse_show(stdout)

  def parse_show({:error, error}), do: {:error, error}

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

  @doc ~S"""

  ## Examples

      iex> Architect.Projects.Commit.parse_count("932\n\n")
      932

  """
  def parse_count({:ok, stdout}), do: parse_count(stdout)

  def parse_count({:error, error}), do: {:error, error}

  def parse_count(stdout) when is_binary(stdout) do
    stdout
    |> String.split("\n")
    |> parse_count()
  end

  def parse_count([line | _]) do
    {count, _} = Integer.parse(line)

    count
  end

  def list_for_ref(dir, ref) do
    {_out, 0} = System.cmd("git", ["checkout", "--force", ref], cd: dir)
    {out, 0} = System.cmd("git", ["log", "--format=#{format()}"], cd: dir)

    parse(out)
  end

  def get_by_sha(dir, sha) do
    {out, 0} = System.cmd("git", ["show", "-s", "--format=#{format()}", sha], cd: dir)

    out
    |> parse_show
  end

  def count_for_branch(dir, branch) do
    {_out, 0} = System.cmd("git", ["checkout", "--force", branch], cd: dir)
    {out, 0} = System.cmd("git", ["rev-list", "--count", branch], cd: dir)

    out
    |> parse_count()
  end

  def count(dir) do
    {out, 0} = System.cmd("git", ["rev-list", "--count", "--all"], cd: dir)

    out
    |> parse_count()
  end
end
