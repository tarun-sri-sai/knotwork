import LoginForm from "./LoginForm";
import { normalizeSearchParams } from "@/lib/search/normalize";
import ErrorAlert from "@/components/ErrorAlert";

const Login = async ({
  searchParams,
}: {
  searchParams: Promise<Record<string, string | string[] | undefined>>;
}) => {
  const resolved = await searchParams;
  const params = normalizeSearchParams(resolved);
  const error = params?.error as string;

  return (
    <div className="flex flex-col gap-4 items-center justify-center p-24">
      <LoginForm />
      {error && <ErrorAlert message={error} />}
    </div>
  );
};

export default Login;
