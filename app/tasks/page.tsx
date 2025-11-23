import { createReadOnlyClient } from "@/lib/supabase/server";
import { redirect } from "next/navigation";

const Tasks = async () => {
  const supabase = await createReadOnlyClient();

  const {
    data: { user },
  } = await supabase.auth.getUser();

  if (!user) {
    redirect("/login");
  }

  return <div>Tasks</div>;
};

export default Tasks;
