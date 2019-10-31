CREATE TABLE IF NOT EXISTS projects (
  id uuid,
  name VARCHAR (256) NOT NULL,
  address VARCHAR(256) NOT NULL,
  ssh_private_key TEXT,
  ssh_host_key TEXT,
  created_at TIMESTAMP WITH TIME ZONE,
  updated_at TIMESTAMP WITH TIME ZONE,
  PRIMARY KEY (id)
);
