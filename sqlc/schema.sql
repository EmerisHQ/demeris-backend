CREATE TABLE IF NOT EXISTS chains (
                                          id serial unique primary key,
                                          enabled boolean default false,
                                          chain_name string not null,
                                          valid_block_thresh string not null,
                                          logo string not null,
                                          display_name string not null,
                                          primary_channel jsonb not null,
                                          denoms jsonb not null,
                                          demeris_addresses text[] not null,
                                          genesis_hash string not null,
                                          node_info jsonb not null,
                                          derivation_path string not null,
                                          unique(chain_name)
)