import { handleResend } from "./actions";
import ErrorMessage from "@/components/ErrorMessage"
import InfoMessage from "@/components/InfoMessage"

const ConfirmEmail = async ({
  searchParams,
}: {
  searchParams: Promise<Record<string, string | string[] | undefined>>;
}) => {
  const resolvedSearchParams = await searchParams;
  const error =
    (Array.isArray(resolvedSearchParams.error)
      ? resolvedSearchParams.error[0]
      : resolvedSearchParams.error) ?? "";

  const info = (Array.isArray(resolvedSearchParams.info)
      ? resolvedSearchParams.info[0]
      : resolvedSearchParams.info) ?? "";

  return (
    <form className="flex flex-col items-center justify-center gap-20">
      <h2 className="text-lg">A confirmation email has been sent to your email address, please head to the home page once confirmed</h2>

      <div className="flex flex-col gap-4">
        <div className="field grid grid-cols-2 gap-4">
          <label htmlFor="email">Missed it? Enter your email to re-send: </label>
          <input id="email" name="email" type="email" />
        </div>

        <div className="field grid grid-cols-2 gap-4">
          <button
            formAction={handleResend}
            type="submit"
            className="text-right"
          >
            Resend email
          </button>
        </div>
      </div>

      {error && <ErrorMessage message={error} />}
      {info && <InfoMessage message={info} />}
    </form>
  );
};

export default ConfirmEmail;
