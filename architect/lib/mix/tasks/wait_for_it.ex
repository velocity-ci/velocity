defmodule Mix.Tasks.WaitForIt do
  use Mix.Task

  @shortdoc "Waits for database from configuration"
  @spec run(any()) :: :ok
  def run(_) do
    repo_config = Application.get_env(:architect, Architect.Repo)
    hostname = to_charlist(repo_config[:hostname])

    IO.puts("testing tcp connectivity to #{hostname}:#{repo_config[:port]}")

    case test_tcp_conn(hostname, repo_config[:port], 1, 10) do
      {:ok} ->
        exit({:shutdown, 0})

      _ ->
        exit({:shutdown, 1})
    end
  end

  defp test_tcp_conn(hostname, port, n, max_tries) when n > max_tries do
    :gen_tcp.connect(hostname, port, [:binary, active: false])
  end

  defp test_tcp_conn(hostname, port, n, max_tries) do
    if n > 1 do
      :timer.sleep(1000)
    end

    IO.write("attempt #{n} - ")

    case :gen_tcp.connect(hostname, port, [:binary, active: false]) do
      {:error, code} ->
        IO.puts(code)
        test_tcp_conn(hostname, port, n + 1, max_tries)
        {:error}

      {:ok, resp} ->
        {:ok, resp}
        IO.puts("success!")
        {:ok}
    end
  end
end
