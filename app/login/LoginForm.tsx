"use client";

import { useState } from "react";
import { login } from "./actions";
import { redirect } from "next/navigation";
import Button from "@/components/Button";
import Input from "@/components/Input";
import ErrorAlert from "@/components/ErrorAlert";

const LoginForm = () => {
  const [email, setEmail] = useState("");
  const [password, setPassword] = useState("");
  const [error, setError] = useState("");
  const [loading, setLoading] = useState(false);

  const handleLogin = async (e: React.FormEvent<HTMLFormElement>) => {
    e.preventDefault();
    setLoading(true);
    const errorMessage = await login(email, password);
    setLoading(false);
    if (errorMessage) {
      setError(errorMessage);
    } else {
      setError("");
      redirect("/tasks");
    }
  };

  return (
    <form onSubmit={handleLogin} className="flex flex-col gap-4">
      <Input
        type="email"
        value={email}
        onChange={(e) => setEmail(e.target.value)}
        placeholder="Email"
      />
      <Input
        type="password"
        value={password}
        onChange={(e) => setPassword(e.target.value)}
        placeholder="Password"
      />
      <Button type="submit" disabled={loading}>{loading ? "Loading..." : "Login"}</Button>
      {error && <ErrorAlert message={error} />}
    </form>
  );
};

export default LoginForm;
