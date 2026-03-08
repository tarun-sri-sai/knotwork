"use server";

import { redirect } from "next/navigation";
import { createClient } from "@/lib/supabase/server";

export const handleSignup = async (formData: FormData) => {
  const email = formData.get("email") as string;
  const password = formData.get("password") as string;
  const confirmPassword = formData.get("confirmPassword") as string;

  if (password !== confirmPassword) {
    redirect(`/signup?error=${encodeURIComponent("Passwords do not match")}`)
  }

  const supabase = await createClient();
  const { error } = await supabase.auth.signUp({ email, password });

  if (error) {
    redirect(`/signup?error=${encodeURIComponent(error.message)}`);
  }

  if ((await supabase.auth.getSession()) === null) {
    redirect(`/confirm-email`);
  }

  redirect("/tasks");
};
