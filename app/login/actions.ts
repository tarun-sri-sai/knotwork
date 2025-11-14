"use server";

import { createClient } from "@/lib/supabase/server";

export const login = async (email: string, password: string) => {
  const supabase = await createClient();
  const { error } = await supabase.auth.signInWithPassword({
    email,
    password,
  });
  return error?.message;
};
