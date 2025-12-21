const ErrorMessage = ({ message }: { message: string }) => {
  return (
    <div className="px-4 py-2 rounded-xl bg-red-500/10 text-red-700">
      {message}
    </div>
  );
};

export default ErrorMessage;
