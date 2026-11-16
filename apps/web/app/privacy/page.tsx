import InfoPageView, { infoMetadata } from "../info-page";

export function generateMetadata() {
  return infoMetadata("privacy");
}

export default function Page() {
  return <InfoPageView slug="privacy" />;
}
