import InfoPageView, { infoMetadata } from "../info-page";

export function generateMetadata() {
  return infoMetadata("about");
}

export default function Page() {
  return <InfoPageView slug="about" />;
}
