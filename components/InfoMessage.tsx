const InfoMessage = ({ message }: { message: string }) => {
  return (
    <div className="px-4 py-2 rounded-xl bg-green-500/10 text-green-700">
      {message}
    </div>
  );
};

export default InfoMessage;
