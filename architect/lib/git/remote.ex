defmodule Git.Repository.Remote do
  require Logger

  def fetch(address, private_key, dir) do
    env = get_environment([], private_key)

    with :ok <- setup_known_host(address),
         %Porcelain.Result{err: nil, out: _, status: 0} <-
           Porcelain.exec("git", ["fetch", "--progress", "--prune"], dir: dir, env: env),
         :ok <- remove_private_key(env),
         :ok <- remove_known_host(address) do
      Logger.debug("Successfully fetched #{address} in #{dir}")
      :ok
    else
      _ ->
        :ok = remove_private_key(env)
        :ok = remove_known_host(address)
        :error
    end
  end

  def verify(address, private_key, dir) do
    env = get_environment([], private_key)

    with :ok <- setup_known_host(address),
         %Porcelain.Result{err: nil, out: _, status: 0} <-
           Porcelain.exec("git", ["ls-remote"], dir: dir, env: env),
         :ok <- remove_private_key(env),
         :ok <- remove_known_host(address) do
      Logger.debug("Successfully verified #{address} in #{dir}")
      :ok
    else
      _ ->
        :ok = remove_private_key(env)
        :ok = remove_known_host(address)
        :error
    end
  end

  defp host_from_address(address) do
    [conn, _path] = String.split(address, ":")
    [_user, host] = String.split(conn, "@")
    host
  end

  defp respect_proxy_environment(env) do
    # TODO
    env
  end

  defp get_environment(env, nil) do
    respect_proxy_environment(env)
  end

  defp get_environment(env, private_key) do
    {:ok, private_key_path} = Temp.open("key", &IO.write(&1, private_key))
    File.chmod(private_key_path, 0o400)
    Logger.debug("wrote a private key to #{private_key_path}")

    [{"GIT_SSH_COMMAND", "ssh -i #{private_key_path}"} | env]
    |> respect_proxy_environment()
  end

  defp setup_known_host("http" <> _address), do: :ok

  defp setup_known_host("git" <> address) do
    known_host = Architect.KnownHosts.get_known_host_by_host!(host_from_address(address))

    case known_host.verified do
      true ->
        Architect.KnownHosts.Node.host_present(known_host.host, known_host.entry)

      false ->
        Logger.warn("#{known_host.host} is not a verified host")
        Architect.KnownHosts.Node.host_absent(known_host.host)
        :error
    end
  end

  defp remove_known_host("http" <> _address), do: :ok

  defp remove_known_host("git" <> address) do
    Architect.KnownHosts.Node.host_absent(host_from_address(address))
  end

  defp remove_private_key(env) do
    Enum.each(env, fn {var, value} ->
      if var == "GIT_SSH_COMMAND" do
        private_key_path =
          value
          |> String.split(" ")
          |> List.last()

        File.rm(private_key_path)
        Logger.debug("removed a private key at #{private_key_path}")
      end
    end)
  end
end
