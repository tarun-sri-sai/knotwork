"use server";

import { redirect } from "next/navigation";

export const handleSignup = async () => {
  redirect("/signup");
};

export const handleLogin = async () => {
  redirect("/login");
};
