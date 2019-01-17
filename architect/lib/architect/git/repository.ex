defmodule Architect.Git.Repository do
  @enforce_keys [:address]
  defstruct [:address, :private_key, :directory]

  defp clone(repo) do
    # repo = %Architect.Git.Repository{
    #   address: "https://github.com/velocity-ci/velocity.git",
    #   private_key: "",
    #   directory:
    #     Path.join(get_workspace, Slugger.slugify("https://github.com/velocity-ci/velocity.git"))
    # }

    init_repo_dir(repo)
    Porcelain.exec("git", ["fetch", "--progress"], dir: repo.directory)
  end

  def get_repo(repo) do
    repo = Map.put(repo, :directory, Path.join(get_workspace(), Slugger.slugify(repo.address)))

    unless File.dir?(Path.join(repo.directory, ".git")) do
      clone(repo)
    end

    repo
  end

  def clean(repo) do
    repo = get_repo(repo)
    Porcelain.exec("git", ["clean", "-fd"], dir: repo.directory)
  end

  def checkout(repo, ref) do
    repo = get_repo(repo)
    Porcelain.exec("git", ["checkout", "--force", ref], dir: repo.directory)
  end

  def describe(repo) do
    repo = get_repo(repo)
    res = Porcelain.exec("git", ["describe", "--always"])
    String.trim(res.out)
  end

  def default_branch(repo) do
    repo = get_repo(repo)
    res = Porcelain.exec("git", ["remote", "show", "origin"], dir: repo.directory)

    String.split(res.out, "\n")
    |> Enum.at(3)
    |> String.split(":")
    |> List.last()
    |> String.trim()
  end

  defp get_workspace do
    {:ok, cwd} = File.cwd()
    Path.join(cwd, "_velocity_data/repositories")
  end

  defp init_repo_dir(repo) do
    unless File.dir?("#{repo.directory}/.git") do
      File.rm_rf(repo.directory)
      File.mkdir_p(repo.directory)
      Porcelain.exec("git", ["init"], dir: repo.directory)
    end

    Porcelain.exec("git", ["remote", "remove", "origin"], dir: repo.directory)
    Porcelain.exec("git", ["remote", "add", "origin", repo.address], dir: repo.directory)
  end
end
