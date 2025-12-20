create table threads (
  id uuid primary key default gen_random_uuid(),
  user_id uuid references auth.users(id) not null,
  parent_id uuid references threads(id),
  completed_at timestamptz default null,
  created_at timestamptz default now(),
  deleted_at timestamptz default null
);

create table comments (
  id uuid primary key default gen_random_uuid(),
  thread_id uuid references threads(id) not null,
  contents text not null,
  created_at timestamptz default now(),
  deleted_at timestamptz default null
);

create index threads_user_id_idx on threads(user_id);
create index threads_parent_id_idx on threads(parent_id);

create index comments_thread_id_idx on comments(thread_id);

alter table threads enable row level security;

create policy "Users can view their own threads"
  on threads for select
  using (auth.uid() = user_id);

create policy "Users can insert their own threads"
  on threads for insert
  with check (auth.uid() = user_id);

create policy "Users can update their own threads"
  on threads for update
  using (auth.uid() = user_id);

alter table comments enable row level security;

create policy "Users can view their own comments"
    on comments for select
    using (exists (
        select 1 from threads
        where threads.id = comments.thread_id
        and threads.user_id = auth.uid()
    ));

create policy "Users can insert comments on their own threads"
    on comments for insert
    with check (exists (
        select 1 from threads
        where threads.id = comments.thread_id
        and threads.user_id = auth.uid()
    ));

create policy "Users can update comments on their own threads"
    on comments for update
    using (exists (
        select 1 from threads
        where threads.id = comments.thread_id
        and threads.user_id = auth.uid()
    ));
