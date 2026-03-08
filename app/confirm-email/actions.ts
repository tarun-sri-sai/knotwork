"use server";

import { createClient } from "@/lib/supabase/server";
import { redirect } from "next/navigation";

export const handleResend = async (formData: FormData) => {
  const email = formData.get("email") as string;

  const supabase = await createClient();

  const { error } = await supabase.auth.resend({ type: "signup", email });

  if (error) {
    redirect(`/confirm-email?error=${encodeURIComponent(error.message)}`);
  }

  redirect(
    `/confirm-email?info=${encodeURIComponent("Sent the confirmation email! Check your inbox and spam")}`,
  );
};
