defmodule Git.Repository.Remote do
  require Logger

  def fetch(address, private_key, known_hosts, dir) do
    env = get_environment([], private_key, known_hosts)

    with {_out, 0} <-
           System.cmd("git", ["fetch", "--progress", "--prune"],
             cd: dir,
             env: env,
             stderr_to_stdout: true
           ),
         :ok <- remove_private_key(address, env),
         :ok <- remove_known_host(address, env) do
      Logger.debug("Successfully fetched #{address} in #{dir}")
      :ok
    else
      _ ->
        :ok = remove_private_key(address, env)
        :ok = remove_known_host(address, env)
        :error
    end
  end

  def verify(address, private_key, known_hosts, dir) do
    env = get_environment([], private_key, known_hosts)

    with {_out, 0} <- System.cmd("git", ["ls-remote"], cd: dir, env: env, stderr_to_stdout: true),
         :ok <- remove_private_key(address, env),
         :ok <- remove_known_host(address, env) do
      Logger.debug("Successfully verified #{address} in #{dir}")
      :ok
    else
      _ ->
        :ok = remove_private_key(address, env)
        :ok = remove_known_host(address, env)
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

  defp get_environment(env, private_key, known_hosts) do
    {:ok, private_key_path} = Temp.open("key", &IO.write(&1, private_key))
    File.chmod(private_key_path, 0o400)
    Logger.debug("wrote a private key to #{private_key_path}")

    {:ok, known_hosts_path} = Temp.open("knownhosts", &IO.write(&1, known_hosts))
    File.chmod(known_hosts_path, 0o600)
    Logger.debug("wrote known hosts to #{known_hosts_path}")

    [
      {"GIT_SSH_COMMAND", "ssh -i #{private_key_path} -o UserKnownHostsFile=#{known_hosts_path}"}
      | env
    ]
    |> respect_proxy_environment()
  end

  # defp setup_known_host("http" <> _address), do: :ok

  # defp setup_known_host("git" <> address) do
  #   known_host = Architect.KnownHosts.get_known_host_by_host!(host_from_address(address))

  #   case known_host.verified do
  #     true ->
  #       Architect.KnownHosts.Node.host_present(known_host.host, known_host.entry)

  #     false ->
  #       Logger.warn("#{known_host.host} is not a verified host")
  #       Architect.KnownHosts.Node.host_absent(known_host.host)
  #       :error
  #   end
  # end

  defp remove_known_host("http" <> _address, env), do: :ok
  defp remove_private_key("http" <> _address, env), do: :ok

  defp remove_known_host("git" <> address, env) do
    Enum.each(env, fn {var, value} ->
      if var == "GIT_SSH_COMMAND" do
        value_parts = String.split(value, " ")
        opts_param_index = Enum.find_index(value_parts, fn x -> x == "-o" end)
        opts = Enum.at(value_parts, opts_param_index + 1)

        opts_parts = String.split(opts, "=")
        ukhf_param_index = Enum.find_index(opts_parts, fn x -> x == "UserKnownHostsFile" end)
        ukhf_path = Enum.at(opts_parts, ukhf_param_index + 1)

        File.rm(ukhf_path)
        Logger.debug("removed known hosts at #{ukhf_path}")
      end
    end)
  end

  defp remove_private_key("git" <> address, env) do
    Enum.each(env, fn {var, value} ->
      if var == "GIT_SSH_COMMAND" do
        value_parts = String.split(value, " ")
        key_param_index = Enum.find_index(value_parts, fn x -> x == "-i" end)
        private_key_path = Enum.at(value_parts, key_param_index + 1)

        File.rm(private_key_path)
        Logger.debug("removed a private key at #{private_key_path}")
      end
    end)
  end
end
