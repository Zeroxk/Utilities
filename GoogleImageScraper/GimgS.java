import java.awt.image.BufferedImage;
import java.io.BufferedReader;
import java.io.File;
import java.io.IOException;
import java.io.InputStreamReader;
import java.net.MalformedURLException;
import java.net.URL;
import java.net.URLConnection;
import java.util.Scanner;

import javax.imageio.ImageIO;

import org.json.*;

public class GimgS {

	public static void main(String[] args) {
		
		//Scanner sc = new Scanner(System.in);
		StringBuilder keyword = new StringBuilder();
		//keyword.append(sc.next());
		//keyword.append("\n");
        keyword.append("Golang");
		System.out.println("Your keyword is: " + keyword.toString());
		
		try {
			StringBuilder query = new StringBuilder();
			query.append("https://ajax.googleapis.com/ajax/services/search/images?v=1.0&safe=active&q=");
			query.append(keyword);
			
			URL url = new URL(query.toString());
			URLConnection conn = url.openConnection();
			
			String line;
			StringBuilder sb = new StringBuilder();
			BufferedReader br = new BufferedReader(new InputStreamReader(conn.getInputStream()));
			
			while( (line = br.readLine()) != null) {
				sb.append(line);
			}
			
			if( sb.length() > 0) {
				System.out.println("Parsing JSON");
				
				JSONObject jo = new JSONObject(sb.toString()).getJSONObject("responseData");											
				
				JSONArray ja = jo.getJSONArray("results");
				System.out.println("Done parsing JSON\n");
				
				System.out.println("Starting image download");
				for (int i = 0; i < ja.length(); i++) {
					JSONObject rs = (JSONObject) ja.get(i);
					System.out.println(rs.getString("url"));
					String name = rs.getString("url");
					
					URL img = new URL(name);
					String imgName = name.substring(name.lastIndexOf("/")+1);
					
					BufferedImage buffImg = ImageIO.read(img);
					
					File f = new File("imgsJava/" + imgName);
					f.mkdirs();
					f.createNewFile();
					String str = imgName.substring(imgName.lastIndexOf(".")+1);
					ImageIO.write(buffImg, imgName.substring(imgName.lastIndexOf(".")+1), f);
					
					System.out.println("Saved to: " + f.getAbsolutePath());
					
				}
				
				System.out.println("Finished");
			}
			
		} catch (MalformedURLException e) {
			e.printStackTrace();
		} catch (IOException e) {
			e.printStackTrace();
		} catch (JSONException e) {
			e.printStackTrace();
		}

	}

}
