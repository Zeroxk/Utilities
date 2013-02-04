import java.awt.image.BufferedImage;
import java.io.ByteArrayOutputStream;
import java.io.File;
import java.io.IOException;
import java.security.MessageDigest;
import java.security.NoSuchAlgorithmException;
import java.util.ArrayList;
import java.util.HashMap;

import javax.imageio.ImageIO;

/**
 * This program detects and deletes duplicate images in a folder by comparing MD5 hashes
 * arg[i] must be absolute path to folder
 * 
 * @author Zeroxk
 *
 */

public class ImgDupDeleter {

	static ArrayList<String> imgExts = new ArrayList<String>();

	public static void main(String[] args) {

		if(args.length < 1) {
			System.out.println("No folders specified");
			System.exit(0);
		}

		imgExts.add(new String("png"));
		imgExts.add(new String("jpg"));

		for (int i = 0; i < args.length; i++) {

			File folder = new File(args[i]);

			if(folder.isDirectory()) {
				checkFolder(folder);
			}else {
				System.out.println(args[i] + " is not a folder");
			}

		}

	}

	/**
	 * Checks a folder for duplicates, does recursive call if the file being worked on is a folder
	 * @param folder Folder to be checked
	 */
	private static void checkFolder(File folder) {

		System.out.println("Processing folder: " + folder.getAbsolutePath());
		File [] images = folder.listFiles();
		int numDupes = 0;
		HashMap<String, File> mapFiles = new HashMap<String, File>();

		for (int j = 0; j < images.length; j++) {
			File currFile = images[j];

			if(currFile.isDirectory()) {
				checkFolder(currFile);
				continue;
			}

			String name = currFile.getName();
			String imgExt = name.substring(name.lastIndexOf(".")+1);

			if(!imgExts.contains(imgExt)) {
				System.out.println(name + " ignored, not an image");
				continue;
			}
			
			System.out.println("Filextension of image is: " + imgExt);

			System.out.println("Processing image: " + currFile.getAbsolutePath());
			if(currFile.isDirectory()) continue;

			byte[] hash = hash(currFile, imgExt);
			if(hash == null) {
				System.out.println(currFile.getAbsolutePath() + " is null, deleted");
				currFile.delete();
				continue;
			}
			String hex = hashToHex(hash);

			if(mapFiles.containsKey(hex) && !mapFiles.get(hex).equals(currFile)) {
				System.out.println(currFile.getAbsolutePath() + " is duplicate of " + mapFiles.get(hex).getAbsolutePath());
				numDupes++;
				currFile.delete();
			}else {
				mapFiles.put(hex, currFile);
			}
		}

		System.out.println("Number of dupes in folder " + folder.getAbsolutePath() + ": " + numDupes);
		System.out.println("Done with folder: " + folder.getAbsolutePath());

	}


	/**
	 * Transforms MD5 hash to hex string
	 * @param hash Digest to be converted
	 * @return Hex string of the hash
	 */
	private static String hashToHex(byte[] hash) {
		String str = "";

		for (int i = 0; i < hash.length; i++) {
			str += Integer.toString( (hash[i] & 0xff) + 0x100, 16).substring(1);
		}

		return str;
	}

	/**
	 * Hashes an image using MD5
	 * @param img Image to be hashed
	 * @return Returns byte array of the hash
	 */
	private static byte[] hash(File img, String imgExt) {

		byte[] hash = null;
		try {


			BufferedImage buffImg = ImageIO.read(img);
			if(buffImg == null) return null;
			ByteArrayOutputStream outputStream = new ByteArrayOutputStream();
			ImageIO.write(buffImg, imgExt, outputStream);
			byte[] data = outputStream.toByteArray();

			System.out.println("Starting MD5 digest");
			MessageDigest md = MessageDigest.getInstance("MD5");
			md.update(data);
			hash = md.digest();

		} catch (IOException e) {
			System.out.println("Error while processing file as image");
			e.printStackTrace();
		} catch (NoSuchAlgorithmException e) {
			System.out.println("Error: could not find MD5 algorithm");
			e.printStackTrace();
		}

		System.out.println("Done with digest of image: " + img.getAbsolutePath() + "\n");
		return hash;

	}

}
