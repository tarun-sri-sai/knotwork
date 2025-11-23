import { createReadOnlyClient } from "@/lib/supabase/server";
import { redirect } from "next/navigation";

const Home = async () => {
  const supabase = await createReadOnlyClient();

  const {
    data: { user },
  } = await supabase.auth.getUser();

  if (!user) {
    redirect("/login");
  }

  redirect("/tasks");
};

export default Home;
